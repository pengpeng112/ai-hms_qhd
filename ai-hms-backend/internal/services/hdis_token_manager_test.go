package services

import (
	"testing"
	"time"
)

func TestShouldRefreshToken(t *testing.T) {
	now := time.Date(2026, 3, 3, 12, 0, 0, 0, time.UTC)

	t.Run("nil expiresAt should not refresh", func(t *testing.T) {
		if shouldRefreshToken(nil, 1800, now) {
			t.Fatalf("expected false, got true")
		}
	})

	t.Run("expires beyond lead window should not refresh", func(t *testing.T) {
		expiresAt := now.Add(2 * time.Hour)
		if shouldRefreshToken(&expiresAt, 1800, now) {
			t.Fatalf("expected false, got true")
		}
	})

	t.Run("expires within lead window should refresh", func(t *testing.T) {
		expiresAt := now.Add(20 * time.Minute)
		if !shouldRefreshToken(&expiresAt, 1800, now) {
			t.Fatalf("expected true, got false")
		}
	})

	t.Run("already expired should refresh", func(t *testing.T) {
		expiresAt := now.Add(-5 * time.Minute)
		if !shouldRefreshToken(&expiresAt, 1800, now) {
			t.Fatalf("expected true, got false")
		}
	})
}

func TestResolveTokenStatus(t *testing.T) {
	now := time.Date(2026, 3, 3, 12, 0, 0, 0, time.UTC)

	t.Run("missing token", func(t *testing.T) {
		if got := resolveTokenStatus(false, nil, 1800, now); got != "MISSING" {
			t.Fatalf("expected MISSING, got %s", got)
		}
	})

	t.Run("token without expiry", func(t *testing.T) {
		if got := resolveTokenStatus(true, nil, 1800, now); got != "UNKNOWN" {
			t.Fatalf("expected UNKNOWN, got %s", got)
		}
	})

	t.Run("valid token", func(t *testing.T) {
		expiresAt := now.Add(90 * time.Minute)
		if got := resolveTokenStatus(true, &expiresAt, 1800, now); got != "VALID" {
			t.Fatalf("expected VALID, got %s", got)
		}
	})

	t.Run("expiring token", func(t *testing.T) {
		expiresAt := now.Add(10 * time.Minute)
		if got := resolveTokenStatus(true, &expiresAt, 1800, now); got != "EXPIRING" {
			t.Fatalf("expected EXPIRING, got %s", got)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		expiresAt := now.Add(-1 * time.Minute)
		if got := resolveTokenStatus(true, &expiresAt, 1800, now); got != "EXPIRED" {
			t.Fatalf("expected EXPIRED, got %s", got)
		}
	})
}
