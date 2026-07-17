package crypto

import (
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := KeyFromString("test-encryption-key")
	ct, err := EncryptString(key, "webhook-secret-value")
	if err != nil {
		t.Fatal(err)
	}
	if ct == "webhook-secret-value" {
		t.Fatal("expected ciphertext, got plaintext")
	}
	pt, err := DecryptString(key, ct)
	if err != nil {
		t.Fatal(err)
	}
	if pt != "webhook-secret-value" {
		t.Fatalf("got %q", pt)
	}
}

func TestDecryptLegacyPlaintext(t *testing.T) {
	key := KeyFromString("test-encryption-key")
	pt, err := DecryptString(key, "legacy-plain-secret")
	if err != nil {
		t.Fatal(err)
	}
	if pt != "legacy-plain-secret" {
		t.Fatalf("got %q", pt)
	}
}
