package his_oracle

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/sijms/go-ora/v2"
)

type Config struct {
	Host     string
	Port     int
	Service  string
	Username string
	Password string
}

func (c Config) DSN() string {
	return fmt.Sprintf("oracle://%s:%s@%s:%d/%s",
		c.Username, c.Password, c.Host, c.Port, c.Service)
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.Host) == "" {
		return errors.New("HIS Oracle host is required")
	}
	if c.Port <= 0 {
		return errors.New("HIS Oracle port is required")
	}
	if strings.TrimSpace(c.Service) == "" {
		return errors.New("HIS Oracle service name is required")
	}
	if strings.TrimSpace(c.Username) == "" {
		return errors.New("HIS Oracle username is required")
	}
	if strings.TrimSpace(c.Password) == "" {
		return errors.New("HIS Oracle password is required")
	}
	return nil
}

type Client struct {
	db  *sql.DB
	cfg Config
}

func NewClient(cfg Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	db, err := sql.Open("oracle", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("HIS Oracle open failed: %w", err)
	}

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &Client{db: db, cfg: cfg}, nil
}

func (c *Client) Ping(ctx context.Context) error {
	if c.db == nil {
		return errors.New("HIS Oracle client not initialized")
	}
	return c.db.PingContext(ctx)
}

func (c *Client) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

const examReportSyncSQL = `
SELECT * FROM (
  SELECT a.exam_no,
         a.patient_id,
         a.visit_id,
         a.name,
         a.sex,
         a.date_of_birth,
         a.exam_class,
         a.exam_sub_class,
         a.performed_by,
         a.req_dept,
         a.req_physician,
         a.req_date_time,
         a.exam_date_time,
         a.report_date_time,
         a.result_status,
         a.study_uid,
         b.description,
         b.impression,
         b.recommendation,
         b.exam_diag,
         b.exam_items AS report_exam_items,
         b.is_abnormal,
         b.use_image,
         b.memo,
         b.reporter,
         b.report_time,
         b.createdate,
         p.id_no,
         v.inp_no,
         v.clinic_no,
         cm.visit_no,
         p.medical_no
    FROM his.exam_master a
    JOIN his.exam_report b ON a.exam_no = b.exam_no
    LEFT JOIN his.pat_master_index p ON p.patient_id = a.patient_id
    LEFT JOIN his.pat_visit v ON v.patient_id = a.patient_id AND v.visit_id = a.visit_id
    LEFT JOIN his.clinic_master cm ON cm.patient_id = a.patient_id
   WHERE a.result_status IN ('3', '4')
     AND b.createdate > :1
   ORDER BY b.createdate ASC, a.exam_no ASC
) WHERE ROWNUM <= :2
`

const examReportByPatientSQL = `
SELECT * FROM (
  SELECT a.exam_no,
         a.patient_id,
         a.visit_id,
         a.name,
         a.sex,
         a.date_of_birth,
         a.exam_class,
         a.exam_sub_class,
         a.performed_by,
         a.req_dept,
         a.req_physician,
         a.req_date_time,
         a.exam_date_time,
         a.report_date_time,
         a.result_status,
         a.study_uid,
         b.description,
         b.impression,
         b.recommendation,
         b.exam_diag,
         b.exam_items AS report_exam_items,
         b.is_abnormal,
         b.use_image,
         b.memo,
         b.reporter,
         b.report_time,
         b.createdate,
         p.id_no,
         v.inp_no,
         v.clinic_no,
         cm.visit_no,
         p.medical_no
    FROM his.exam_master a
    JOIN his.exam_report b ON a.exam_no = b.exam_no
    LEFT JOIN his.pat_master_index p ON p.patient_id = a.patient_id
    LEFT JOIN his.pat_visit v ON v.patient_id = a.patient_id AND v.visit_id = a.visit_id
    LEFT JOIN his.clinic_master cm ON cm.patient_id = a.patient_id
   WHERE a.patient_id = :1
     AND a.result_status IN ('3', '4')
     AND b.createdate > :2
   ORDER BY b.createdate ASC, a.exam_no ASC
) WHERE ROWNUM <= :3
`

type QueryExamReportsParams struct {
	CursorTime time.Time
	BatchSize  int
	PatientID  string
}

func (c *Client) QueryExamReports(ctx context.Context, params QueryExamReportsParams) ([]HisExamRow, error) {
	if c.db == nil {
		return nil, errors.New("HIS Oracle client not initialized")
	}

	var rows *sql.Rows
	var err error

	if strings.TrimSpace(params.PatientID) != "" {
		rows, err = c.db.QueryContext(ctx, examReportByPatientSQL,
			params.PatientID, params.CursorTime, params.BatchSize)
	} else {
		rows, err = c.db.QueryContext(ctx, examReportSyncSQL,
			params.CursorTime, params.BatchSize)
	}
	if err != nil {
		return nil, fmt.Errorf("HIS Oracle query failed: %w", err)
	}
	defer rows.Close()

	var results []HisExamRow
	for rows.Next() {
		var r HisExamRow
		var visitID sql.NullInt64
		var reqDateTime, examDateTime, reportDateTime, reportTime, createDate sql.NullTime
		var dateOfBirth sql.NullTime

		err := rows.Scan(
			&r.ExamNo, &r.PatientID, &visitID,
			&r.Name, &r.Sex, &dateOfBirth,
			&r.ExamClass, &r.ExamSubClass, &r.PerformedBy, &r.ReqDept, &r.ReqPhysician,
			&reqDateTime, &examDateTime, &reportDateTime,
			&r.ResultStatus, &r.StudyUID,
			&r.Description, &r.Impression, &r.Recommendation, &r.ExamDiag,
			&r.ReportExamItems, &r.IsAbnormal, &r.UseImage, &r.Memo,
			&r.Reporter, &reportTime, &createDate,
			&r.IDNo,
			&r.InpNo,
			&r.ClinicNo,
			&r.VisitNo,
			&r.MedicalNo,
		)
		if err != nil {
			return results, fmt.Errorf("HIS Oracle row scan failed: %w", err)
		}

		if visitID.Valid {
			r.VisitID = &visitID.Int64
		}
		if dateOfBirth.Valid {
			r.DateOfBirth = &dateOfBirth.Time
		}
		if reqDateTime.Valid {
			r.ReqDateTime = &reqDateTime.Time
		}
		if examDateTime.Valid {
			r.ExamDateTime = &examDateTime.Time
		}
		if reportDateTime.Valid {
			r.ReportDateTime = &reportDateTime.Time
		}
		if reportTime.Valid {
			r.ReportTime = &reportTime.Time
		}
		if createDate.Valid {
			r.CreateDate = &createDate.Time
		}

		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return results, fmt.Errorf("HIS Oracle rows iteration error: %w", err)
	}

	return results, nil
}

func (c *Client) TestConnection(ctx context.Context) (time.Duration, error) {
	start := time.Now()
	if err := c.Ping(ctx); err != nil {
		return 0, err
	}
	return time.Since(start), nil
}

func (c *Client) FindPatientIDByIDNo(ctx context.Context, idNo string) (string, error) {
	if c.db == nil {
		return "", errors.New("HIS Oracle client not initialized")
	}
	rows, err := c.db.QueryContext(ctx,
		`SELECT patient_id FROM his.pat_master_index WHERE id_no = :1 AND ROWNUM <= 1`, idNo)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	if rows.Next() {
		var pid string
		if err := rows.Scan(&pid); err != nil {
			return "", err
		}
		return pid, nil
	}
	return "", nil
}

type UnmatchedPatientRow struct {
	PatientID string
	NameVal   string
	ExamCnt   int
}

type UnmatchedPatientsParams struct {
	Page     int
	PageSize int
	Keyword  string
}

type UnmatchedPatientsResult struct {
	Items []UnmatchedPatientRow
	Total int
}

func (c *Client) QueryUnmatchedPatients(ctx context.Context, params UnmatchedPatientsParams) (*UnmatchedPatientsResult, error) {
	if c.db == nil {
		return nil, errors.New("HIS Oracle client not initialized")
	}
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}

	keyword := strings.TrimSpace(params.Keyword)

	var whereFilter string
	var countArgs []interface{}

	if keyword != "" {
		kw := "%" + keyword + "%"
		countArgs = append(countArgs, kw)
		whereFilter = ` AND (a.name LIKE :1 OR TO_CHAR(a.patient_id) LIKE :1)`
	}

	countSQL := `SELECT COUNT(*) FROM (
		SELECT a.patient_id FROM his.exam_master a
		WHERE a.result_status IN ('3','4')` + whereFilter + `
		GROUP BY a.patient_id
	)`

	var total int
	if err := c.db.QueryRowContext(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count query failed: %w", err)
	}

	fetchSize := params.PageSize * 3
	startRow := (params.Page - 1) * params.PageSize
	endRow := startRow + fetchSize

	posEnd := len(countArgs) + 1
	posStart := posEnd + 1
	dataArgs := append(countArgs, endRow, startRow)

	dataSQL := fmt.Sprintf(`SELECT * FROM (
		SELECT inner.*, ROWNUM rn FROM (
			SELECT a.patient_id,
			       MAX(a.name) AS name_val,
			       COUNT(*) AS exam_cnt
			  FROM his.exam_master a
			 WHERE a.result_status IN ('3','4')%s
			 GROUP BY a.patient_id
			 ORDER BY exam_cnt DESC
		) inner WHERE ROWNUM <= :%d
	) WHERE rn > :%d`, whereFilter, posEnd, posStart)

	rows, err := c.db.QueryContext(ctx, dataSQL, dataArgs...)
	if err != nil {
		return nil, fmt.Errorf("unmatched query failed: %w", err)
	}
	defer rows.Close()

	var items []UnmatchedPatientRow
	for rows.Next() {
		var r UnmatchedPatientRow
		var rn int
		if err := rows.Scan(&r.PatientID, &r.NameVal, &r.ExamCnt, &rn); err != nil {
			return nil, err
		}
		items = append(items, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return &UnmatchedPatientsResult{
		Items: items,
		Total: total,
	}, nil
}

func (c *Client) QueryExamItems(ctx context.Context, examNos []string) ([]HisExamItemRow, error) {
	if c.db == nil {
		return nil, errors.New("HIS Oracle client not initialized")
	}
	if len(examNos) == 0 {
		return nil, nil
	}

	const batchSize = 500
	var all []HisExamItemRow

	for i := 0; i < len(examNos); i += batchSize {
		end := i + batchSize
		if end > len(examNos) {
			end = len(examNos)
		}
		batch := examNos[i:end]

		placeholders := make([]string, len(batch))
		args := make([]interface{}, len(batch))
		for j, en := range batch {
			placeholders[j] = fmt.Sprintf(":%d", j+1)
			args[j] = en
		}

		sql := fmt.Sprintf(`SELECT exam_no, exam_item, exam_item_code, exam_item_no
			FROM his.exam_items WHERE exam_no IN (%s) ORDER BY exam_no, exam_item_no`,
			strings.Join(placeholders, ","))

		rows, err := c.db.QueryContext(ctx, sql, args...)
		if err != nil {
			return all, fmt.Errorf("exam items batch query failed: %w", err)
		}

		for rows.Next() {
			var r HisExamItemRow
			if err := rows.Scan(&r.ExamNo, &r.ExamItem, &r.ExamItemCode, &r.ExamItemNo); err != nil {
				rows.Close()
				return all, err
			}
			all = append(all, r)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return all, err
		}
		rows.Close()
	}

	return all, nil
}
