package clinicalsafety

import (
	"errors"
	"testing"
	"time"
)

func TestVersionETagRoundTrip(t *testing.T) {
	for _, v := range []int{0, 1, 7, 999999} {
		etag := VersionETag(v)
		got, ok := ParseVersionETag(etag)
		if !ok || got != v {
			t.Fatalf("version %d: etag=%q parsed=(%d,%v)", v, etag, got, ok)
		}
	}
	// 兼容客户端原样回传（含 W/ 与引号）
	if got, ok := ParseVersionETag(`W/"v42"`); !ok || got != 42 {
		t.Fatalf("parse W/\"v42\" => (%d,%v)", got, ok)
	}
}

func TestTimeETagRoundTrip(t *testing.T) {
	orig := time.Date(2026, 6, 10, 8, 30, 0, 123456789, time.UTC)
	etag := TimeETag(orig)
	got, ok := ParseTimeETag(etag)
	if !ok || !got.Equal(orig) {
		t.Fatalf("time etag=%q parsed=(%v,%v) want %v", etag, got, ok, orig)
	}
}

func TestParseETagInvalid(t *testing.T) {
	cases := []string{"", "v7", `"v7`, `W/"`, `W/"x7"`, `W/"vabc"`, `W/"tabc"`, "garbage"}
	for _, c := range cases {
		if _, ok := ParseVersionETag(c); ok {
			t.Errorf("ParseVersionETag(%q) unexpectedly ok", c)
		}
		if _, ok := ParseTimeETag(c); ok && c != `W/"t"` {
			// "t" with empty number also invalid; just ensure no panic and false
			t.Errorf("ParseTimeETag(%q) unexpectedly ok", c)
		}
	}
}

func TestVersionAndTimeTagsDoNotCross(t *testing.T) {
	if _, ok := ParseTimeETag(VersionETag(5)); ok {
		t.Error("version etag parsed as time etag")
	}
	if _, ok := ParseVersionETag(TimeETag(time.Now())); ok {
		t.Error("time etag parsed as version etag")
	}
}

func TestIsConflict(t *testing.T) {
	err := &VersionConflictError{Entity: "Treatment", ID: 1001}
	if !IsConflict(err) {
		t.Fatal("IsConflict should be true for *VersionConflictError")
	}
	if IsConflict(errors.New("other")) {
		t.Fatal("IsConflict should be false for unrelated error")
	}
	if IsConflict(nil) {
		t.Fatal("IsConflict(nil) should be false")
	}
	// 包裹后仍可识别
	wrapped := errors.Join(errors.New("ctx"), err)
	if !IsConflict(wrapped) {
		t.Fatal("IsConflict should detect wrapped conflict")
	}
}
