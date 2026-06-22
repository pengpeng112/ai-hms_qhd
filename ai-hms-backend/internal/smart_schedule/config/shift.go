package config

import (
	_ "embed"
	"encoding/json"
)

// 班次配置（C3）：go:embed 内嵌 JSON，院方增删班次模式（early/late/long/short…）重构建即生效。

//go:embed shifts.json
var shiftsJSON []byte

// ShiftDef 一种人员班次定义
type ShiftDef struct {
	Code    string `json:"code"`
	Name    string `json:"name"`
	Start   string `json:"start"`
	End     string `json:"end"`
	Enabled bool   `json:"enabled"`
}

var loadedShifts []ShiftDef

func init() {
	_ = json.Unmarshal(shiftsJSON, &loadedShifts)
}

// EnabledShifts 返回启用的班次（供前端下拉与校验）
func EnabledShifts() []ShiftDef {
	out := make([]ShiftDef, 0, len(loadedShifts))
	for _, s := range loadedShifts {
		if s.Enabled {
			out = append(out, s)
		}
	}
	return out
}

// ValidShiftCode 班次码是否为启用班次
func ValidShiftCode(code string) bool {
	for _, s := range loadedShifts {
		if s.Enabled && s.Code == code {
			return true
		}
	}
	return false
}
