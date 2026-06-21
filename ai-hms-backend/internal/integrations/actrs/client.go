package actrs

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
	"sync"
	"time"
)

const defaultTimeout = 10 * time.Second

type Client struct {
	baseURL string
	cfg     Config
	http    *http.Client
	mu      sync.Mutex
	token   string
}

func NewClient(cfg Config) *Client {
	to := time.Duration(cfg.TimeoutSec) * time.Second
	if to <= 0 {
		to = defaultTimeout
	}
	return &Client{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		cfg:     cfg,
		http:    &http.Client{Timeout: to},
	}
}

func (c *Client) login(ctx context.Context) (string, error) {
	body, _ := json.Marshal(loginRequest{Username: c.cfg.Username, Password: c.cfg.Password})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/auth/login", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("actrs login failed: status %d", resp.StatusCode)
	}
	var lr loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return "", err
	}
	if lr.AccessToken == "" {
		return "", errors.New("actrs login: empty token")
	}
	return lr.AccessToken, nil
}

func (c *Client) do(ctx context.Context, method, path string, contentType string, body []byte, bodyFn func() (io.Reader, string, error)) ([]byte, error) {
	attempt := func() (*http.Response, error) {
		c.mu.Lock()
		tok := c.token
		c.mu.Unlock()
		if tok == "" {
			t, err := c.login(ctx)
			if err != nil {
				return nil, err
			}
			c.mu.Lock()
			c.token = t
			c.mu.Unlock()
			tok = t
		}
		var rdr io.Reader
		ct := contentType
		if bodyFn != nil {
			r, c2, err := bodyFn()
			if err != nil {
				return nil, err
			}
			rdr, ct = r, c2
		} else if body != nil {
			rdr = bytes.NewReader(body)
		}
		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, rdr)
		if err != nil {
			return nil, err
		}
		if ct != "" {
			req.Header.Set("Content-Type", ct)
		}
		req.Header.Set("Authorization", "Bearer "+tok)
		return c.http.Do(req)
	}

	resp, err := attempt()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 401 {
		resp.Body.Close()
		c.mu.Lock()
		c.token = ""
		c.mu.Unlock()
		resp, err = attempt()
		if err != nil {
			return nil, err
		}
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("actrs %s %s -> status %d: %s", method, path, resp.StatusCode, string(data))
	}
	return data, nil
}

func (c *Client) UpsertPatient(ctx context.Context, in PatientCreate) (*PatientOut, error) {
	body, _ := json.Marshal(in)
	data, err := c.do(ctx, http.MethodPost, "/patients", "application/json", body, nil)
	if err != nil {
		return nil, err
	}
	var out PatientOut
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ListXrays(ctx context.Context, actrsPatientID int64) ([]XrayOut, error) {
	data, err := c.do(ctx, http.MethodGet, fmt.Sprintf("/patients/%d/xrays", actrsPatientID), "", nil, nil)
	if err != nil {
		return nil, err
	}
	var out []XrayOut
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) AnalyzeXray(ctx context.Context, actrsPatientID int64, filename string, file io.Reader) (*XrayOut, error) {
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("actrs read upload: %w", err)
	}
	bodyFn := func() (io.Reader, string, error) {
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		fw, err := w.CreateFormFile("file", filename)
		if err != nil {
			return nil, "", err
		}
		if _, err := fw.Write(fileBytes); err != nil {
			return nil, "", err
		}
		w.Close()
		return &buf, w.FormDataContentType(), nil
	}
	data, err := c.do(ctx, http.MethodPost, fmt.Sprintf("/patients/%d/xrays", actrsPatientID), "", nil, bodyFn)
	if err != nil {
		return nil, err
	}
	var out XrayOut
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ApplyCorrection(ctx context.Context, xrayID int64, in CorrectionRequest) (*XrayOut, error) {
	body, _ := json.Marshal(in)
	data, err := c.do(ctx, http.MethodPatch, fmt.Sprintf("/xrays/%d/correction", xrayID), "application/json", body, nil)
	if err != nil {
		return nil, err
	}
	var out XrayOut
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
