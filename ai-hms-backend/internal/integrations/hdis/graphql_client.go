package hdis

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

// GraphQLClient HDIS GraphQL 客户端
// 端点格式示例：https://hdis.ingatek.com:7778/api/python/pygql
type GraphQLClient struct {
	endpoint   string
	token      string
	httpClient *http.Client
}

// NewGraphQLClient 创建 GraphQL 客户端
func NewGraphQLClient(endpoint, token string, timeout time.Duration) *GraphQLClient {
	if timeout <= 0 {
		timeout = defaultHTTPTimeout
	}

	return &GraphQLClient{
		endpoint:   strings.TrimSpace(endpoint),
		token:      strings.TrimSpace(token),
		httpClient: &http.Client{Timeout: timeout},
	}
}

// Query 执行 GraphQL 查询（multipart/form-data）
func (c *GraphQLClient) Query(ctx context.Context, query string) (map[string]json.RawMessage, error) {
	if strings.TrimSpace(c.endpoint) == "" {
		return nil, errors.New("hdis graphql endpoint is empty")
	}
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("graphql query is required")
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("query", query); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("hdis graphql http status %d: %s", resp.StatusCode, string(respBody))
	}

	trimmed := strings.TrimSpace(string(respBody))
	if trimmed == "" || trimmed == "null" {
		return map[string]json.RawMessage{}, nil
	}

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal graphql response failed: %w", err)
	}

	if rawErrors, ok := payload["errors"]; ok {
		var gqlErrors []GraphQLError
		if err := json.Unmarshal(rawErrors, &gqlErrors); err == nil && len(gqlErrors) > 0 {
			return nil, fmt.Errorf("hdis graphql failed: %s", gqlErrors[0].Message)
		}
		return nil, errors.New("hdis graphql failed with unknown error")
	}

	return payload, nil
}

// GetExaminationItems 查询指定报告的明细项
func (c *GraphQLClient) GetExaminationItems(ctx context.Context, examinationID int) ([]GraphQLExaminationItem, error) {
	if examinationID <= 0 {
		return nil, errors.New("examinationID must be greater than 0")
	}

	query := buildExaminationItemsQuery(examinationID)
	payload, err := c.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	// 兼容 HDIS 两种返回：
	// 1) { "ExaminationItem": [...] }
	// 2) { "data": { "ExaminationItem": [...] } }
	if raw, ok := payload["ExaminationItem"]; ok {
		var items []GraphQLExaminationItem
		if err := json.Unmarshal(raw, &items); err != nil {
			return nil, err
		}
		return items, nil
	}

	if rawData, ok := payload["data"]; ok {
		var wrapped struct {
			ExaminationItem []GraphQLExaminationItem `json:"ExaminationItem"`
		}
		if err := json.Unmarshal(rawData, &wrapped); err != nil {
			return nil, err
		}
		return wrapped.ExaminationItem, nil
	}

	return []GraphQLExaminationItem{}, nil
}

// GetRecords 查询患者关键指标记录
func (c *GraphQLClient) GetRecords(ctx context.Context, patientID int, page int, pageSize int) ([]GraphQLRecord, error) {
	if patientID <= 0 {
		return nil, errors.New("patientID must be greater than 0")
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}

	query := buildRecordsQuery(patientID, page, pageSize)
	payload, err := c.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	if raw, ok := payload["Record"]; ok {
		var records []GraphQLRecord
		if err := json.Unmarshal(raw, &records); err != nil {
			return nil, err
		}
		return records, nil
	}

	if rawData, ok := payload["data"]; ok {
		var wrapped struct {
			Record []GraphQLRecord `json:"Record"`
		}
		if err := json.Unmarshal(rawData, &wrapped); err != nil {
			return nil, err
		}
		return wrapped.Record, nil
	}

	return []GraphQLRecord{}, nil
}

// GetExaminations 查询患者 Examination 报告头
func (c *GraphQLClient) GetExaminations(ctx context.Context, patientID int) ([]GraphQLExamination, error) {
	if patientID <= 0 {
		return nil, errors.New("patientID must be greater than 0")
	}

	query := buildExaminationsQuery(patientID)
	payload, err := c.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	if raw, ok := payload["Examination"]; ok {
		var examinations []GraphQLExamination
		if err := json.Unmarshal(raw, &examinations); err != nil {
			return nil, err
		}
		return examinations, nil
	}

	if rawData, ok := payload["data"]; ok {
		var wrapped struct {
			Examination []GraphQLExamination `json:"Examination"`
		}
		if err := json.Unmarshal(rawData, &wrapped); err != nil {
			return nil, err
		}
		return wrapped.Examination, nil
	}

	return []GraphQLExamination{}, nil
}

func buildExaminationItemsQuery(examinationID int) string {
	params := fmt.Sprintf("{'examinationId':%d,'sort':'Sort ascend'}", examinationID)
	return fmt.Sprintf(
		`{ExaminationItem(parameters:"%s"){Id,TenantId,ExaminationId,ItemName,ItemCode,Result,Unit,Reference,ResultSign,LastModifyTime}}`,
		params,
	)
}

func buildRecordsQuery(patientID int, page int, pageSize int) string {
	params := fmt.Sprintf("{'pageSize':%d,'page':%d,'sort':'TestTime descend','patientId':%d}", pageSize, page, patientID)
	return fmt.Sprintf(
		`{Record(parameters:"%s"){Id,TenantId,PatientId,IndexName,IndexCode,Result,Unit,Reference,ResultSign,TestTime,EvaluationResult}}`,
		params,
	)
}

func buildExaminationsQuery(patientID int) string {
	params := fmt.Sprintf("{'patientId':%d,'sort':'Name descend'}", patientID)
	return fmt.Sprintf(
		`{Examination(parameters:"%s"){Id,PatientId,Name,Type,ResultTime,TestNO,LastModifyTime}}`,
		params,
	)
}
