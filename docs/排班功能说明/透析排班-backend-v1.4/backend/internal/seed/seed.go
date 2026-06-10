// Package seed 写入演示数据,便于端到端验证排班生成。
package seed

import (
	"gorm.io/gorm"

	"github.com/sdsph/dialysis-scheduling/internal/model"
	"github.com/sdsph/dialysis-scheduling/internal/sched"
)

func p64(v int64) *int64 { return &v }
func i16(v int16) *int16 { return &v }

// Demo 在空库(该租户尚无班次)时写入一套演示数据,返回描述。
func Demo(g *gorm.DB, tenant int64) (string, error) {
	var shiftCount int64
	g.Model(&model.Shift{}).Where(`"TenantId" = ?`, tenant).Count(&shiftCount)
	if shiftCount > 0 {
		return "已存在数据,跳过 seed", nil
	}

	err := g.Transaction(func(tx *gorm.DB) error {
		// 1) 分区 A/B/C
		wardA := &model.Ward{BaseModel: bm(tenant), Name: "A区(门诊)", ZoneType: sched.ZoneA, Sort: 1}
		wardB := &model.Ward{BaseModel: bm(tenant), Name: "B区(住院)", ZoneType: sched.ZoneB, Sort: 2}
		wardC := &model.Ward{BaseModel: bm(tenant), Name: "C区(全警戒)", ZoneType: sched.ZoneC, Sort: 3}
		for _, w := range []*model.Ward{wardA, wardB, wardC} {
			if err := tx.Create(w).Error; err != nil {
				return err
			}
		}

		// 2) 班次 上午/下午/晚
		mor := &model.Shift{BaseModel: bm(tenant), Name: "上午班", ShiftCode: "MORNING", StartTime: "07:30", EndTime: "11:30", Sort: 1}
		aft := &model.Shift{BaseModel: bm(tenant), Name: "下午班", ShiftCode: "AFTERNOON", StartTime: "12:30", EndTime: "16:30", Sort: 2}
		night := &model.Shift{BaseModel: bm(tenant), Name: "晚班", ShiftCode: "NIGHT", StartTime: "17:30", EndTime: "21:30", Sort: 3}
		for _, s := range []*model.Shift{mor, aft, night} {
			if err := tx.Create(s).Error; err != nil {
				return err
			}
		}

		// 3) 机器:A 区 6 台 HD + 2 台 HDF;B 区 4 台 HD;C 区 2 台 HDF + 2 台 CRRT
		var machines []*model.Machine
		pos := 0
		add := func(ward int64, code, mt string) {
			pos++
			machines = append(machines, &model.Machine{
				BaseModel: bm(tenant), WardId: ward, Code: code, Name: code, MachineType: mt, PositionIndex: pos,
			})
		}
		pos = 0
		for i := 1; i <= 6; i++ {
			add(wardA.Id, "A-HD-"+itoa(i), sched.MachineHD)
		}
		add(wardA.Id, "A-HDF-1", sched.MachineHDF)
		add(wardA.Id, "A-HDF-2", sched.MachineHDF)
		pos = 0
		for i := 1; i <= 4; i++ {
			add(wardB.Id, "B-HD-"+itoa(i), sched.MachineHD)
		}
		pos = 0
		add(wardC.Id, "C-HDF-1", sched.MachineHDF)
		add(wardC.Id, "C-HDF-2", sched.MachineHDF)
		add(wardC.Id, "C-CRRT-1", sched.MachineCRRT)
		add(wardC.Id, "C-CRRT-2", sched.MachineCRRT)
		if err := tx.CreateInBatches(machines, 50).Error; err != nil {
			return err
		}

		// 4) 模板头
		tpl := &model.ScheduleTemplate{BaseModel: bm(tenant), Name: "标准周模板", Scope: "ALL", IsActive: true}
		if err := tx.Create(tpl).Error; err != nil {
			return err
		}

		// 5) 病人:A 区上午班,涵盖五种频率与一个 HDF 病人
		type pt struct {
			id     int64
			name   string
			gender string
			freq   int16
			hdf    bool
			hdfWd  *int16
		}
		pts := []pt{
			{1001, "张伟", "男", sched.FreqMonWedFri, false, nil},
			{1002, "李芳", "女", sched.FreqMonWedFri, true, i16(3)}, // 一三五,HDF 日=周三
			{1003, "王强", "男", sched.FreqTueThuSat, false, nil},
			{1004, "刘洋", "男", sched.FreqTwoPerWk, false, nil},
			{1005, "陈静", "女", sched.FreqOnePerWk, false, nil},
			{1006, "赵磊", "男", sched.FreqMonWedFri, false, nil},
			{1007, "孙丽", "女", sched.FreqTueThuSat, true, i16(2)}, // 二四六,HDF 日=周二
		}
		for _, p := range pts {
			if err := tx.Create(&model.Patient{Id: p.id, TenantId: tenant, Name: p.name, Gender: p.gender}).Error; err != nil {
				return err
			}
			prof := &model.PatientProfile{
				BaseModel: bm(tenant), PatientId: p.id, ZoneTag: sched.ZoneA,
				HomeWardId: p64(wardA.Id), FreqPattern: p.freq, ShiftId: p64(mor.Id),
				DefaultMode: sched.ModeHD, HdfEnabled: p.hdf, HdfWeekday: p.hdfWd,
			}
			if err := tx.Create(prof).Error; err != nil {
				return err
			}
			item := &model.ScheduleTemplateItem{
				BaseModel: bm(tenant), TemplateId: tpl.Id, PatientId: p.id, ZoneTag: sched.ZoneA,
				WardId: p64(wardA.Id), ShiftId: p64(mor.Id), FreqPattern: p.freq, DefaultMode: sched.ModeHD,
				HdfEnabled: p.hdf, HdfWeekday: p.hdfWd,
			}
			if err := tx.Create(item).Error; err != nil {
				return err
			}
		}

		// 6) 配置:奇偶周基准周一
		setting := &model.TenantSetting{BaseModel: bm(tenant), SettingKey: sched.CfgAnchorMonday, SettingValue: "2025-01-06"}
		return tx.Create(setting).Error
	})
	if err != nil {
		return "", err
	}
	return "演示数据写入成功:3 区 / 3 班 / 14 机 / 7 病人 / 1 模板", nil
}

func bm(tenant int64) model.BaseModel { return model.BaseModel{TenantId: tenant} }

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var b [20]byte
	i := len(b)
	for v > 0 {
		i--
		b[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
