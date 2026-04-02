package hdis

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultHTTPTimeout = 15 * time.Second

// WebcmdClient HDIS webcmd 客户端
// 端点格式示例：https://hdis.ingatek.com:7777/pyapi/webcmd
type WebcmdClient struct {
	endpoint   string
	token      string
	httpClient *http.Client
}

// NewWebcmdClient 创建 webcmd 客户端
func NewWebcmdClient(endpoint, token string, timeout time.Duration) *WebcmdClient {
	if timeout <= 0 {
		timeout = defaultHTTPTimeout
	}

	return &WebcmdClient{
		endpoint:   strings.TrimSpace(endpoint),
		token:      strings.TrimSpace(token),
		httpClient: &http.Client{Timeout: timeout},
	}
}

// Call 执行 webcmd 调用
func (c *WebcmdClient) Call(ctx context.Context, function string, params map[string]any) (json.RawMessage, error) {
	if strings.TrimSpace(c.endpoint) == "" {
		return nil, errors.New("hdis webcmd endpoint is empty")
	}
	if strings.TrimSpace(function) == "" {
		return nil, errors.New("webcmd function is required")
	}
	if params == nil {
		params = map[string]any{}
	}

	payload := map[string]any{
		"function": function,
		"params": map[string]any{
			"params": params,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
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
		return nil, fmt.Errorf("hdis webcmd http status %d: %s", resp.StatusCode, string(respBody))
	}

	var webcmdResp WebcmdResponse
	if err := json.Unmarshal(respBody, &webcmdResp); err != nil {
		return nil, fmt.Errorf("unmarshal webcmd response failed: %w", err)
	}

	if !webcmdResp.Succeeded || webcmdResp.ErrorCode != 0 {
		return nil, fmt.Errorf("hdis webcmd failed: code=%d, message=%s", webcmdResp.ErrorCode, webcmdResp.Message)
	}

	return webcmdResp.Data, nil
}

// GetExaminationList 获取患者检验报告头列表
func (c *WebcmdClient) GetExaminationList(ctx context.Context, patientIDs []int, types []string, names []string) ([]WebcmdExamination, error) {
	if len(patientIDs) == 0 {
		return nil, errors.New("patientIDs is required")
	}

	params := map[string]any{
		"patientIds": patientIDs,
	}
	if len(types) > 0 {
		params["types"] = types
	}
	if len(names) > 0 {
		params["names"] = names
	}

	data, err := c.Call(ctx, "lis_service.get_examination_list", params)
	if err != nil {
		return nil, err
	}

	exams, err := ParseWebcmdExaminationList(data)
	if err != nil {
		return nil, err
	}

	return exams, nil
}
