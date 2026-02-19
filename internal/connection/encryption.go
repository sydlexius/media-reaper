package connection

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

var (
	ErrMasterKeyRequired  = errors.New("master key is required")
	ErrMasterKeyLength    = errors.New("master key must be 32 bytes (64 hex characters)")
	ErrCiphertextTooShort = errors.New("ciphertext too short")
)

// Encryptor provides AES-256-GCM encryption and decryption for API keys.
type Encryptor struct {
	gcm cipher.AEAD
}

// NewEncryptor creates an Encryptor from a hex-encoded 32-byte master key.
func NewEncryptor(hexKey string) (*Encryptor, error) {
	if hexKey == "" {
		return nil, ErrMasterKeyRequired
	}

	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("decoding master key: %w", err)
	}

	if len(key) != 32 {
		return nil, ErrMasterKeyLength
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("creating cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("creating GCM: %w", err)
	}

	return &Encryptor{gcm: gcm}, nil
}

// Encrypt encrypts plaintext and returns hex-encoded nonce+ciphertext.
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, e.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generating nonce: %w", err)
	}

	sealed := e.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(sealed), nil
}

// Decrypt decrypts hex-encoded nonce+ciphertext and returns plaintext.
func (e *Encryptor) Decrypt(hexCiphertext string) (string, error) {
	data, err := hex.DecodeString(hexCiphertext)
	if err != nil {
		return "", fmt.Errorf("decoding ciphertext: %w", err)
	}

	nonceSize := e.gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrCiphertextTooShort
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := e.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypting: %w", err)
	}

	return string(plaintext), nil
}
