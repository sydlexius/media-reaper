package connection

import (
	"crypto/rand"
	"encoding/hex"
	"testing"
)

func generateTestKey(t *testing.T) string {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("generating test key: %v", err)
	}
	return hex.EncodeToString(key)
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	enc, err := NewEncryptor(generateTestKey(t))
	if err != nil {
		t.Fatalf("creating encryptor: %v", err)
	}

	tests := []string{
		"my-secret-api-key",
		"short",
		"a-very-long-api-key-that-is-much-longer-than-typical-keys-1234567890abcdef",
		"special chars: !@#$%^&*()_+-=[]{}|;':\",./<>?",
	}

	for _, plaintext := range tests {
		ciphertext, err := enc.Encrypt(plaintext)
		if err != nil {
			t.Fatalf("encrypting %q: %v", plaintext, err)
		}

		decrypted, err := enc.Decrypt(ciphertext)
		if err != nil {
			t.Fatalf("decrypting %q: %v", plaintext, err)
		}

		if decrypted != plaintext {
			t.Errorf("round-trip failed: got %q, want %q", decrypted, plaintext)
		}
	}
}

func TestEncryptProducesDifferentCiphertexts(t *testing.T) {
	enc, err := NewEncryptor(generateTestKey(t))
	if err != nil {
		t.Fatalf("creating encryptor: %v", err)
	}

	plaintext := "same-api-key"
	ct1, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("first encrypt: %v", err)
	}

	ct2, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("second encrypt: %v", err)
	}

	if ct1 == ct2 {
		t.Error("two encryptions of the same plaintext produced identical ciphertexts")
	}

	// Both should still decrypt to the same value
	d1, _ := enc.Decrypt(ct1)
	d2, _ := enc.Decrypt(ct2)
	if d1 != d2 || d1 != plaintext {
		t.Error("different ciphertexts did not decrypt to the same plaintext")
	}
}

func TestEncryptDecryptEmptyString(t *testing.T) {
	enc, err := NewEncryptor(generateTestKey(t))
	if err != nil {
		t.Fatalf("creating encryptor: %v", err)
	}

	ciphertext, err := enc.Encrypt("")
	if err != nil {
		t.Fatalf("encrypting empty string: %v", err)
	}

	decrypted, err := enc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decrypting empty string: %v", err)
	}

	if decrypted != "" {
		t.Errorf("expected empty string, got %q", decrypted)
	}
}

func TestNewEncryptorEmptyKey(t *testing.T) {
	_, err := NewEncryptor("")
	if err != ErrMasterKeyRequired {
		t.Errorf("expected ErrMasterKeyRequired, got %v", err)
	}
}

func TestNewEncryptorWrongLength(t *testing.T) {
	// 16 bytes (32 hex chars) instead of 32 bytes
	shortKey := hex.EncodeToString(make([]byte, 16))
	_, err := NewEncryptor(shortKey)
	if err != ErrMasterKeyLength {
		t.Errorf("expected ErrMasterKeyLength, got %v", err)
	}

	// 64 bytes (128 hex chars)
	longKey := hex.EncodeToString(make([]byte, 64))
	_, err = NewEncryptor(longKey)
	if err != ErrMasterKeyLength {
		t.Errorf("expected ErrMasterKeyLength, got %v", err)
	}
}

func TestNewEncryptorInvalidHex(t *testing.T) {
	_, err := NewEncryptor("not-valid-hex-string-at-all-needs-to-be-sixty-four-chars-long!!")
	if err == nil {
		t.Error("expected error for invalid hex key")
	}
}

func TestDecryptTamperedCiphertext(t *testing.T) {
	enc, err := NewEncryptor(generateTestKey(t))
	if err != nil {
		t.Fatalf("creating encryptor: %v", err)
	}

	ciphertext, err := enc.Encrypt("api-key-to-tamper")
	if err != nil {
		t.Fatalf("encrypting: %v", err)
	}

	// Tamper with the ciphertext by flipping a byte near the end
	tampered := []byte(ciphertext)
	if tampered[len(tampered)-1] == 'a' {
		tampered[len(tampered)-1] = 'b'
	} else {
		tampered[len(tampered)-1] = 'a'
	}

	_, err = enc.Decrypt(string(tampered))
	if err == nil {
		t.Error("expected error when decrypting tampered ciphertext")
	}
}

func TestDecryptInvalidHex(t *testing.T) {
	enc, err := NewEncryptor(generateTestKey(t))
	if err != nil {
		t.Fatalf("creating encryptor: %v", err)
	}

	_, err = enc.Decrypt("not-valid-hex")
	if err == nil {
		t.Error("expected error for invalid hex ciphertext")
	}
}

func TestDecryptTooShort(t *testing.T) {
	enc, err := NewEncryptor(generateTestKey(t))
	if err != nil {
		t.Fatalf("creating encryptor: %v", err)
	}

	_, err = enc.Decrypt(hex.EncodeToString([]byte("short")))
	if err != ErrCiphertextTooShort {
		t.Errorf("expected ErrCiphertextTooShort, got %v", err)
	}
}

func TestDecryptWrongKey(t *testing.T) {
	enc1, err := NewEncryptor(generateTestKey(t))
	if err != nil {
		t.Fatalf("creating encryptor 1: %v", err)
	}

	enc2, err := NewEncryptor(generateTestKey(t))
	if err != nil {
		t.Fatalf("creating encryptor 2: %v", err)
	}

	ciphertext, err := enc1.Encrypt("secret-api-key")
	if err != nil {
		t.Fatalf("encrypting: %v", err)
	}

	_, err = enc2.Decrypt(ciphertext)
	if err == nil {
		t.Error("expected error when decrypting with wrong key")
	}
}
