package utils

import "testing"

func TestSecretBoxEncryptDecrypt(t *testing.T) {
	box := NewSecretBox("test-secret")
	plaintext := "sensitive-value-123"

	ciphertext, err := box.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}
	if ciphertext == plaintext {
		t.Fatalf("ciphertext should not equal plaintext")
	}

	got, err := box.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}
	if got != plaintext {
		t.Fatalf("expected %q, got %q", plaintext, got)
	}
}

func TestSecretBoxDecryptWithDifferentKeyShouldFail(t *testing.T) {
	origin := NewSecretBox("origin-secret")
	other := NewSecretBox("other-secret")

	ciphertext, err := origin.Encrypt("token-abc")
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	if _, err := other.Decrypt(ciphertext); err == nil {
		t.Fatalf("expected decrypt error with different key")
	}
}
