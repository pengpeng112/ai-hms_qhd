package services

// PatientBasicInfoResponse 患者基本信息档案响应
type PatientBasicInfoResponse struct {
	// 身份核心信息
	PersonalInfo PatientBasicPersonal `json:"personalInfo"`

	// 医疗登记信息
	MedicalInfo PatientBasicMedical `json:"medicalInfo"`

	// 生命体征与社会信息
	VitalSocialInfo PatientBasicVitalSocial `json:"vitalSocialInfo"`

	// 联系信息
	ContactInfo PatientBasicContact `json:"contactInfo"`
}

// PatientBasicPersonal 身份核心信息
type PatientBasicPersonal struct {
	Name        string  `json:"name"`                // 患者姓名
	Pinyin      *string `json:"pinyin,omitempty"`    // 姓名拼音
	Birthday    *string `json:"birthday,omitempty"`  // 出生日期
	Age         int     `json:"age"`                 // 当前年龄
	Gender      string  `json:"gender"`              // 性别：男/女
	Ethnicity   *string `json:"ethnicity,omitempty"` // 民族
	IDType      string  `json:"idType"`              // 身份证件类型
	IDNumber    string  `json:"idNumber"`            // 身份证号
	PatientType string  `json:"patientType"`         // 患者类型：门诊/住院
}

// PatientBasicMedical 医疗登记信息
type PatientBasicMedical struct {
	VisitCategory         *string `json:"visitCategory,omitempty"`         // 就诊类别
	AdmissionNo           *string `json:"admissionNo,omitempty"`           // 住院号
	VisitNo               *string `json:"visitNo,omitempty"`               // 就诊号
	MedicalRecordNo       *string `json:"medicalRecordNo,omitempty"`       // 病历号
	InsuranceNo           *string `json:"insuranceNo,omitempty"`           // 医保号
	HdisPatientId         *int    `json:"hdisPatientId,omitempty"`         // HDIS/LIS 患者数字 ID
	InsuranceType         string  `json:"insuranceType"`                   // 医保类型
	DialysisNo            *string `json:"dialysisNo,omitempty"`            // 透析号
	DoctorName            string  `json:"doctorName"`                      // 经治医生
	NurseName             *string `json:"nurseName,omitempty"`             // 责任护士
	FirstDialysisDate     *string `json:"firstDialysisDate,omitempty"`     // 首次透析日期
	FirstHospitalDate     *string `json:"firstHospitalDate,omitempty"`     // 首次在本院透析日期
	FirstDialysisHospital *string `json:"firstDialysisHospital,omitempty"` // 首次透析医院
	CurrentDialysisAge    *string `json:"currentDialysisAge,omitempty"`    // 当前透析龄
}

// PatientBasicVitalSocial 生命体征与社会信息
type PatientBasicVitalSocial struct {
	Height         *string `json:"height,omitempty"`         // 身高 (cm)
	DryWeight      float64 `json:"dryWeight"`                // 干体重 (kg)
	ABOBloodType   *string `json:"aboBloodType,omitempty"`   // ABO血型
	RhBloodType    *string `json:"rhBloodType,omitempty"`    // Rh血型
	EducationLevel *string `json:"educationLevel,omitempty"` // 文化程度
	Occupation     *string `json:"occupation,omitempty"`     // 职业
	MaritalStatus  *string `json:"maritalStatus,omitempty"`  // 婚姻状况
	Workplace      *string `json:"workplace,omitempty"`      // 工作单位
}

// PatientBasicContact 联系信息
type PatientBasicContact struct {
	Phone        *string `json:"phone,omitempty"`        // 手机号码
	Wechat       *string `json:"wechat,omitempty"`       // 微信号
	Landline     *string `json:"landline,omitempty"`     // 固定电话
	Address      *string `json:"address,omitempty"`      // 地址
	District     *string `json:"district,omitempty"`     // 区域
	ContactName  *string `json:"contactName,omitempty"`  // 紧急联系人
	ContactPhone *string `json:"contactPhone,omitempty"` // 紧急联系电话
}

// PatientBasicInfoRequest 更新患者基本信息档案请求
type PatientBasicInfoRequest struct {
	// 身份核心信息
	PersonalInfo PatientBasicPersonalRequest `json:"personalInfo"`

	// 医疗登记信息
	MedicalInfo PatientBasicMedicalRequest `json:"medicalInfo"`

	// 生命体征与社会信息
	VitalSocialInfo PatientBasicVitalSocialRequest `json:"vitalSocialInfo"`

	// 联系信息
	ContactInfo PatientBasicContactRequest `json:"contactInfo"`
}

// PatientBasicPersonalRequest 身份核心信息请求
type PatientBasicPersonalRequest struct {
	Name        *string `json:"name"`        // 患者姓名
	Pinyin      *string `json:"pinyin"`      // 姓名拼音
	Birthday    *string `json:"birthday"`    // 出生日期
	Gender      *string `json:"gender"`      // 性别：男/女
	Ethnicity   *string `json:"ethnicity"`   // 民族
	IDType      *string `json:"idType"`      // 身份证件类型
	IDNumber    *string `json:"idNumber"`    // 身份证号
	PatientType *string `json:"patientType"` // 患者类型：门诊/住院
}

// PatientBasicMedicalRequest 医疗登记信息请求
type PatientBasicMedicalRequest struct {
	VisitCategory         *string `json:"visitCategory"`         // 就诊类别
	AdmissionNo           *string `json:"admissionNo"`           // 住院号
	VisitNo               *string `json:"visitNo"`               // 就诊号
	MedicalRecordNo       *string `json:"medicalRecordNo"`       // 病历号
	InsuranceNo           *string `json:"insuranceNo"`           // 医保号
	HdisPatientId         *int    `json:"hdisPatientId"`         // HDIS/LIS 患者数字 ID
	InsuranceType         *string `json:"insuranceType"`         // 医保类型
	DialysisNo            *string `json:"dialysisNo"`            // 透析号
	DoctorName            *string `json:"doctorName"`            // 经治医生
	NurseName             *string `json:"nurseName"`             // 责任护士
	FirstDialysisDate     *string `json:"firstDialysisDate"`     // 首次透析日期
	FirstHospitalDate     *string `json:"firstHospitalDate"`     // 首次在本院透析日期
	FirstDialysisHospital *string `json:"firstDialysisHospital"` // 首次透析医院
}

// PatientBasicVitalSocialRequest 生命体征与社会信息请求
type PatientBasicVitalSocialRequest struct {
	Height         *string  `json:"height"`         // 身高
	DryWeight      *float64 `json:"dryWeight"`      // 干体重
	ABOBloodType   *string  `json:"aboBloodType"`   // ABO血型
	RhBloodType    *string  `json:"rhBloodType"`    // Rh血型
	EducationLevel *string  `json:"educationLevel"` // 文化程度
	Occupation     *string  `json:"occupation"`     // 职业
	MaritalStatus  *string  `json:"maritalStatus"`  // 婚姻状况
	Workplace      *string  `json:"workplace"`      // 工作单位
}

// PatientBasicContactRequest 联系信息请求
type PatientBasicContactRequest struct {
	Phone        *string `json:"phone"`        // 手机号码
	Wechat       *string `json:"wechat"`       // 微信号
	Landline     *string `json:"landline"`     // 固定电话
	Address      *string `json:"address"`      // 地址
	District     *string `json:"district"`     // 区域
	ContactName  *string `json:"contactName"`  // 紧急联系人
	ContactPhone *string `json:"contactPhone"` // 紧急联系电话
}
