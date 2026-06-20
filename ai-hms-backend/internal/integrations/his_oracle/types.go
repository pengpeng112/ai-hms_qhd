package his_oracle

import "time"

type HisExamRow struct {
	ExamNo          string     `json:"examNo"`
	PatientID       string     `json:"patientId"`
	VisitID         *int64     `json:"visitId,omitempty"`
	Name            *string    `json:"name,omitempty"`
	Sex             *string    `json:"sex,omitempty"`
	DateOfBirth     *time.Time `json:"dateOfBirth,omitempty"`
	ExamClass       *string    `json:"examClass,omitempty"`
	ExamSubClass    *string    `json:"examSubClass,omitempty"`
	PerformedBy     *string    `json:"performedBy,omitempty"`
	ReqDept         *string    `json:"reqDept,omitempty"`
	ReqPhysician    *string    `json:"reqPhysician,omitempty"`
	ReqDateTime     *time.Time `json:"reqDateTime,omitempty"`
	ExamDateTime    *time.Time `json:"examDateTime,omitempty"`
	ReportDateTime  *time.Time `json:"reportDateTime,omitempty"`
	ResultStatus    *string    `json:"resultStatus,omitempty"`
	StudyUID        *string    `json:"studyUid,omitempty"`
	Description     *string    `json:"description,omitempty"`
	Impression      *string    `json:"impression,omitempty"`
	Recommendation  *string    `json:"recommendation,omitempty"`
	ExamDiag        *string    `json:"examDiag,omitempty"`
	ReportExamItems *string    `json:"reportExamItems,omitempty"`
	IsAbnormal      *string    `json:"isAbnormal,omitempty"`
	UseImage        *string    `json:"useImage,omitempty"`
	Memo            *string    `json:"memo,omitempty"`
	Reporter        *string    `json:"reporter,omitempty"`
	ReportTime      *time.Time `json:"reportTime,omitempty"`
	CreateDate      *time.Time `json:"createDate,omitempty"`
	ItemNames       *string    `json:"itemNames,omitempty"`
	// 患者主索引字段
	IDNo *string `json:"idNo,omitempty"`
	// 住院就诊字段
	InpNo      *string `json:"inpNo,omitempty"`
	ClinicNo   *string `json:"clinicNo,omitempty"`
	MedicalNo  *string `json:"medicalNo,omitempty"`
	VisitNo    *string `json:"visitNo,omitempty"`
}

type HisExamItemRow struct {
	ExamNo       string  `json:"examNo"`
	ExamItem     *string `json:"examItem,omitempty"`
	ExamItemCode *string `json:"examItemCode,omitempty"`
	ExamItemNo   *string `json:"examItemNo,omitempty"`
}
