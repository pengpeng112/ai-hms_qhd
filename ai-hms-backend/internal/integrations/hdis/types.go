package hdis

import "encoding/json"

// WebcmdResponse webcmd 响应格式
// 文档示例：{"Succeeded":true,"ErrorCode":0,"Data":[...],"Message":"成功"}
type WebcmdResponse struct {
	Succeeded bool            `json:"Succeeded"`
	ErrorCode int             `json:"ErrorCode"`
	Data      json.RawMessage `json:"Data"`
	Message   string          `json:"Message"`
}

// WebcmdGroup 检验报告分组
type WebcmdGroup struct {
	GroupName  string              `json:"groupName"`
	GroupValue []WebcmdExamination `json:"groupValue"`
	AllLength  int                 `json:"allLength"`
}

// WebcmdExamination LIS 报告头
// 字段基于 docs/检验报告_API_文档.md
// 只声明同步流程会用到的字段，其他字段可按需扩展。
type WebcmdExamination struct {
	ID                    int     `json:"Id"`
	PatientID             *int    `json:"PatientId"`
	Name                  string  `json:"Name"`
	Type                  string  `json:"Type"`
	ResultTime            string  `json:"ResultTime"`
	Time                  string  `json:"Time"`
	LastModifyTime        string  `json:"LastModifyTime"`
	SyncTime              string  `json:"SyncTime"`
	TestNO                string  `json:"TestNO"`
	Priority              *int    `json:"Priority"`
	ClinicalDiagnosisDesc *string `json:"ClinicalDiagnosisDesc"`
	Specimen              string  `json:"Specimen"`
	ApplyDept             string  `json:"ApplyDept"`
	DealDept              string  `json:"DealDept"`
	SpecimenReceivedTime  string  `json:"SpecimenReceivedTime"`
	SpecimenSampleTime    string  `json:"SpecimenSampleTime"`
	ApplyTime             string  `json:"ApplyTime"`
	ApplyUserName         string  `json:"ApplyUserName"`
	ResultStatus          *int    `json:"ResultStatus"`
	ResultRPTTime         string  `json:"ResultRPTTime"`
}

// DataCategory 外部数据分类
type DataCategory string

const (
	DataCategoryLab  DataCategory = "LAB"
	DataCategoryExam DataCategory = "EXAM"
)

// GraphQLError GraphQL 错误结构
type GraphQLError struct {
	Message string `json:"message"`
}

// GraphQLExaminationItem HDIS GraphQL 明细项
// 注意：部分字段在不同租户可能返回 number/null，这里使用 any 做兼容转换。
type GraphQLExaminationItem struct {
	ID            int `json:"Id"`
	ExaminationID int `json:"ExaminationId"`

	ItemName any `json:"ItemName"`
	ItemCode any `json:"ItemCode"`
	Result   any `json:"Result"`
	Unit     any `json:"Unit"`

	Reference  any `json:"Reference"`
	ResultSign any `json:"ResultSign"`

	// 兼容字段：部分租户仍返回旧字段名。
	RefRange     any `json:"RefRange"`
	AbnormalFlag any `json:"AbnormalFlag"`

	LastModifyTime any `json:"LastModifyTime"`
}

// GraphQLRecord HDIS 关键指标记录
// 注意：不同租户数值类型可能混用 number/string/null，统一用 any 做兼容。
type GraphQLRecord struct {
	ID any `json:"Id"`

	PatientID any `json:"PatientId"`
	TenantID  any `json:"TenantId"`

	IndexName any `json:"IndexName"`
	IndexCode any `json:"IndexCode"`

	Result     any `json:"Result"`
	Unit       any `json:"Unit"`
	Reference  any `json:"Reference"`
	ResultSign any `json:"ResultSign"`

	TestTime any `json:"TestTime"`

	EvaluationResult any `json:"EvaluationResult"`
}

// GraphQLExamination HDIS Examination 报告头
// 注意：不同租户可能返回 number/string/null 混用，统一用 any 做兼容。
type GraphQLExamination struct {
	ID        any `json:"Id"`
	PatientID any `json:"PatientId"`

	Name any `json:"Name"`
	Type any `json:"Type"`

	ResultTime     any `json:"ResultTime"`
	LastModifyTime any `json:"LastModifyTime"`
	TestNO         any `json:"TestNO"`
}
