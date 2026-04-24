package services

type PatientBasicInfoResponse struct {
	PersonalInfo    PatientBasicPersonal    `json:"personalInfo"`
	MedicalInfo     PatientBasicMedical     `json:"medicalInfo"`
	VitalSocialInfo PatientBasicVitalSocial `json:"vitalSocialInfo"`
	ContactInfo     PatientBasicContact     `json:"contactInfo"`
}

type PatientBasicPersonal struct {
	Name        string  `json:"name"`
	Pinyin      *string `json:"pinyin,omitempty"`
	Birthday    *string `json:"birthday,omitempty"`
	Age         int     `json:"age"`
	Gender      string  `json:"gender"`
	Ethnicity   *string `json:"ethnicity,omitempty"`
	IDType      string  `json:"idType"`
	IDNumber    string  `json:"idNumber"`
	PatientType string  `json:"patientType"`
}

type PatientBasicMedical struct {
	VisitCategory         *string `json:"visitCategory,omitempty"`
	AdmissionNo           *string `json:"admissionNo,omitempty"`
	VisitNo               *string `json:"visitNo,omitempty"`
	MedicalRecordNo       *string `json:"medicalRecordNo,omitempty"`
	InsuranceNo           *string `json:"insuranceNo,omitempty"`
	HdisPatientId         *int    `json:"hdisPatientId,omitempty"`
	InsuranceType         string  `json:"insuranceType"`
	DialysisNo            *string `json:"dialysisNo,omitempty"`
	DoctorName            string  `json:"doctorName"`
	NurseName             *string `json:"nurseName,omitempty"`
	FirstDialysisDate     *string `json:"firstDialysisDate,omitempty"`
	FirstHospitalDate     *string `json:"firstHospitalDate,omitempty"`
	FirstDialysisHospital *string `json:"firstDialysisHospital,omitempty"`
	CurrentDialysisAge    *string `json:"currentDialysisAge,omitempty"`
}

type PatientBasicVitalSocial struct {
	Height         *string `json:"height,omitempty"`
	DryWeight      float64 `json:"dryWeight"`
	ABOBloodType   *string `json:"aboBloodType,omitempty"`
	RhBloodType    *string `json:"rhBloodType,omitempty"`
	EducationLevel *string `json:"educationLevel,omitempty"`
	Occupation     *string `json:"occupation,omitempty"`
	MaritalStatus  *string `json:"maritalStatus,omitempty"`
	Workplace      *string `json:"workplace,omitempty"`
}

type PatientBasicContact struct {
	Phone        *string `json:"phone,omitempty"`
	Wechat       *string `json:"wechat,omitempty"`
	Landline     *string `json:"landline,omitempty"`
	Address      *string `json:"address,omitempty"`
	District     *string `json:"district,omitempty"`
	ContactName  *string `json:"contactName,omitempty"`
	ContactPhone *string `json:"contactPhone,omitempty"`
}

type PatientBasicInfoRequest struct {
	PersonalInfo    PatientBasicPersonalRequest    `json:"personalInfo"`
	MedicalInfo     PatientBasicMedicalRequest     `json:"medicalInfo"`
	VitalSocialInfo PatientBasicVitalSocialRequest `json:"vitalSocialInfo"`
	ContactInfo     PatientBasicContactRequest     `json:"contactInfo"`
}

type PatientBasicPersonalRequest struct {
	Name        *string `json:"name"`
	Pinyin      *string `json:"pinyin"`
	Birthday    *string `json:"birthday"`
	Gender      *string `json:"gender"`
	Ethnicity   *string `json:"ethnicity"`
	IDType      *string `json:"idType"`
	IDNumber    *string `json:"idNumber"`
	PatientType *string `json:"patientType"`
}

type PatientBasicMedicalRequest struct {
	VisitCategory         *string `json:"visitCategory"`
	AdmissionNo           *string `json:"admissionNo"`
	VisitNo               *string `json:"visitNo"`
	MedicalRecordNo       *string `json:"medicalRecordNo"`
	InsuranceNo           *string `json:"insuranceNo"`
	HdisPatientId         *int    `json:"hdisPatientId"`
	InsuranceType         *string `json:"insuranceType"`
	DialysisNo            *string `json:"dialysisNo"`
	DoctorName            *string `json:"doctorName"`
	NurseName             *string `json:"nurseName"`
	FirstDialysisDate     *string `json:"firstDialysisDate"`
	FirstHospitalDate     *string `json:"firstHospitalDate"`
	FirstDialysisHospital *string `json:"firstDialysisHospital"`
}

type PatientBasicVitalSocialRequest struct {
	Height         *string  `json:"height"`
	DryWeight      *float64 `json:"dryWeight"`
	ABOBloodType   *string  `json:"aboBloodType"`
	RhBloodType    *string  `json:"rhBloodType"`
	EducationLevel *string  `json:"educationLevel"`
	Occupation     *string  `json:"occupation"`
	MaritalStatus  *string  `json:"maritalStatus"`
	Workplace      *string  `json:"workplace"`
}

type PatientBasicContactRequest struct {
	Phone        *string `json:"phone"`
	Wechat       *string `json:"wechat"`
	Landline     *string `json:"landline"`
	Address      *string `json:"address"`
	District     *string `json:"district"`
	ContactName  *string `json:"contactName"`
	ContactPhone *string `json:"contactPhone"`
}
