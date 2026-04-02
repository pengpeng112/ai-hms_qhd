package hdis

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"
)

func buildTestJWT(t *testing.T, payload map[string]any) string {
	t.Helper()
	header := map[string]any{"alg": "none", "typ": "JWT"}
	headerBytes, err := json.Marshal(header)
	if err != nil {
		t.Fatalf("marshal header: %v", err)
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	return base64.RawURLEncoding.EncodeToString(headerBytes) + "." +
		base64.RawURLEncoding.EncodeToString(payloadBytes) + "."
}

func TestParseJWTStringClaim(t *testing.T) {
	token := buildTestJWT(t, map[string]any{
		"organ_id": "2",
		"tenant":   "abc",
	})

	if got := parseJWTStringClaim(token, "organ_id"); got != "2" {
		t.Fatalf("organ_id mismatch, got %q", got)
	}
	if got := parseJWTStringClaim(token, "tenant"); got != "abc" {
		t.Fatalf("tenant mismatch, got %q", got)
	}
}

func TestParseJWTStringClaim_NumberClaim(t *testing.T) {
	token := buildTestJWT(t, map[string]any{
		"organ_id": 4,
	})
	if got := parseJWTStringClaim(token, "organ_id"); got != "4" {
		t.Fatalf("expected numeric claim converted to string, got %q", got)
	}
}

func TestParseJWTExp(t *testing.T) {
	exp := time.Now().Add(30 * time.Minute).Unix()
	token := buildTestJWT(t, map[string]any{
		"exp": exp,
	})
	got := parseJWTExp(token)
	if got == nil {
		t.Fatal("expected exp to be parsed")
	}
	if got.Unix() != exp {
		t.Fatalf("exp mismatch, got %d want %d", got.Unix(), exp)
	}
}

func TestExtractAccessTokenFromURL_Fragment(t *testing.T) {
	href := "https://hdis.ingatek.com:7778#access_token=token123&id_token=abc&token_type=Bearer"
	token, err := extractAccessTokenFromURL(href)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "token123" {
		t.Fatalf("token mismatch, got %q", token)
	}
}

func TestExtractAccessTokenFromURL_Query(t *testing.T) {
	href := "https://hdis.ingatek.com:7778/callback?access_token=query_token&state=xyz"
	token, err := extractAccessTokenFromURL(href)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "query_token" {
		t.Fatalf("token mismatch, got %q", token)
	}
}

func TestExtractAccessTokenFromURL_NotFound(t *testing.T) {
	href := "https://hdis.ingatek.com:7778/#/register/list/1"
	_, err := extractAccessTokenFromURL(href)
	if err == nil {
		t.Fatal("expected error when token not found")
	}
}
