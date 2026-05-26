package crypto

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"

	secondaryports "github.com/nlypage/applycation/backend/internal/ports/secondary"
)

const algorithm = "aes-256-gcm"

type Encrypter struct {
	aead cipher.AEAD
}

var _ secondaryports.Encrypter = (*Encrypter)(nil)

func NewEncrypter(appSecret string) (*Encrypter, error) {
	if appSecret == "" {
		return nil, errors.New("app secret is empty")
	}

	key := sha256.Sum256([]byte(appSecret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}

	return &Encrypter{aead: aead}, nil
}

func (e *Encrypter) Encrypt(ctx context.Context, plaintext []byte) ([]byte, []byte, error) {
	_ = ctx

	nonce := make([]byte, e.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := e.aead.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

func (e *Encrypter) Decrypt(ctx context.Context, ciphertext []byte, nonce []byte) ([]byte, error) {
	_ = ctx

	if len(nonce) != e.aead.NonceSize() {
		return nil, fmt.Errorf("invalid nonce size: got %d, want %d", len(nonce), e.aead.NonceSize())
	}

	plaintext, err := e.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt payload: %w", err)
	}
	return plaintext, nil
}

func (e *Encrypter) Algorithm() string {
	return algorithm
}
