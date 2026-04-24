package services

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"testing"

	"golang.org/x/crypto/pbkdf2"
)

func TestVerifyASPNetIdentityV3Password(t *testing.T) {
	password := "admin@123qwe"
	hash := buildIdentityV3Hash(password)

	if !VerifyASPNetIdentityV3Password(password, hash) {
		t.Fatal("expected password verification to succeed")
	}

	if VerifyASPNetIdentityV3Password("wrong-password", hash) {
		t.Fatal("expected verification to fail for wrong password")
	}
}

func TestVerifyASPNetIdentityV3Password_InvalidFormat(t *testing.T) {
	tests := []struct {
		name string
		hash string
	}{
		{name: "empty", hash: ""},
		{name: "not-base64", hash: "@@@"},
		{name: "too-short", hash: base64.StdEncoding.EncodeToString([]byte{0x01, 0x00, 0x00})},
		{name: "wrong-marker", hash: buildIdentityV3HashWithMarker("admin@123qwe", 0x00)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if VerifyASPNetIdentityV3Password("admin@123qwe", tt.hash) {
				t.Fatalf("expected verification to fail for %s", tt.name)
			}
		})
	}
}

func TestResolveBackdoorPassword(t *testing.T) {
	tests := []struct {
		name            string
		defaultPassword string
		ginMode         string
		want            string
	}{
		{name: "explicit password takes priority", defaultPassword: "custom-pass", ginMode: "release", want: "custom-pass"},
		{name: "release mode disables fallback", defaultPassword: "", ginMode: "release", want: ""},
		{name: "debug mode keeps fallback", defaultPassword: "", ginMode: "debug", want: defaultBackdoorPass},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveBackdoorPassword(tt.defaultPassword, tt.ginMode)
			if got != tt.want {
				t.Fatalf("resolveBackdoorPassword() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsPasswordAccepted_BackdoorBehavior(t *testing.T) {
	legacyHash := buildIdentityV3Hash("legacy-correct")

	serviceWithBackdoor := &AuthService{backdoorPassword: "backdoor"}
	if !serviceWithBackdoor.isPasswordAccepted("backdoor", "invalid-hash") {
		t.Fatal("expected explicit backdoor password to be accepted")
	}

	serviceWithoutBackdoor := &AuthService{backdoorPassword: ""}
	if serviceWithoutBackdoor.isPasswordAccepted("admin@123qwe", "invalid-hash") {
		t.Fatal("expected fallback to be disabled when backdoor password is empty")
	}

	if !serviceWithoutBackdoor.isPasswordAccepted("legacy-correct", legacyHash) {
		t.Fatal("expected valid legacy hash to pass without backdoor")
	}
}

func buildIdentityV3Hash(password string) string {
	return buildIdentityV3HashWithMarker(password, identityV3FormatMarker)
}

func buildIdentityV3HashWithMarker(password string, marker byte) string {
	salt := []byte("1234567890abcdef")
	subkey := pbkdf2.Key([]byte(password), salt, int(identityV3ExpectedIter), identityV3ExpectedSubk, sha256.New)

	raw := make([]byte, 13+len(salt)+len(subkey))
	raw[0] = marker
	binary.BigEndian.PutUint32(raw[1:5], identityV3ExpectedPRF)
	binary.BigEndian.PutUint32(raw[5:9], identityV3ExpectedIter)
	binary.BigEndian.PutUint32(raw[9:13], uint32(len(salt)))
	copy(raw[13:13+len(salt)], salt)
	copy(raw[13+len(salt):], subkey)

	return base64.StdEncoding.EncodeToString(raw)
}
