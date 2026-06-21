package actrs

type Config struct {
	BaseURL    string
	Username   string
	Password   string
	TimeoutSec int
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken string `json:"access_token"`
}

type PatientCreate struct {
	DialysisID string `json:"dialysis_id"`
	Name       string `json:"name"`
}

type PatientOut struct {
	ID         int64  `json:"id"`
	DialysisID string `json:"dialysis_id"`
	Name       string `json:"name"`
}

type XrayOut struct {
	ID               int64    `json:"id"`
	CTR              *float64 `json:"ctr"`
	ACTR             *float64 `json:"actr"`
	ACTR1            *float64 `json:"actr1"`
	ACTR2            *float64 `json:"actr2"`
	ACTRNorm         *float64 `json:"actr_norm"`
	HeartWidth       *int     `json:"heart_width"`
	LungWidth        *int     `json:"lung_width"`
	TiltAngle        *float64 `json:"tilt_angle"`
	QCPaAp           string   `json:"qc_pa_ap"`
	QCPass           bool     `json:"qc_pass"`
	QCWarnings       string   `json:"qc_warnings"`
	ModelVersion     string   `json:"model_version"`
	ImagePath        string   `json:"image_path"`
	OverlayPath      string   `json:"overlay_path"`
	MaskPath         string   `json:"mask_path"`
	DoctorCorrection *float64 `json:"doctor_correction"`
	Notes            string   `json:"notes"`
	AnalysisDate     string   `json:"analysis_date"`
}

type CorrectionRequest struct {
	DoctorCorrection float64 `json:"doctor_correction"`
	Notes            string  `json:"notes,omitempty"`
}
