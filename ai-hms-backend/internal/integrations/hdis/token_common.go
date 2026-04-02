package hdis

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// TokenRefreshResponse 刷新 token 响应
type TokenRefreshResponse struct {
	AccessToken string
	ExpiresAt   time.Time
	OrganID     string
}

func parseJWTExp(token string) *time.Time {
	payload, err := parseJWTPayload(token)
	if err != nil {
		return nil
	}

	expVal, ok := payload["exp"]
	if !ok {
		return nil
	}

	var exp int64
	switch v := expVal.(type) {
	case float64:
		exp = int64(v)
	case int64:
		exp = v
	case string:
		parsed, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		if err != nil {
			return nil
		}
		exp = parsed
	default:
		return nil
	}

	if exp <= 0 {
		return nil
	}

	t := time.Unix(exp, 0)
	return &t
}

func parseJWTStringClaim(token, claim string) string {
	payload, err := parseJWTPayload(token)
	if err != nil {
		return ""
	}
	val, ok := payload[claim]
	if !ok {
		return ""
	}
	switch v := val.(type) {
	case string:
		return strings.TrimSpace(v)
	case float64:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}

func parseJWTPayload(token string) (map[string]any, error) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid jwt")
	}

	raw := parts[1]
	raw = strings.ReplaceAll(raw, "-", "+")
	raw = strings.ReplaceAll(raw, "_", "/")
	if mod := len(raw) % 4; mod != 0 {
		raw += strings.Repeat("=", 4-mod)
	}

	payloadBytes, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, err
	}

	var payload map[string]any
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}
