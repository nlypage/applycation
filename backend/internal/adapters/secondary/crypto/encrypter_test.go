package crypto

import (
	"bytes"
	"context"
	"testing"
)

func TestNewEncrypter_EmptySecret(t *testing.T) {
	t.Parallel()

	_, err := NewEncrypter("")
	if err == nil {
		t.Fatal("expected error for empty secret")
	}
}

func TestEncrypter_EncryptDecryptRoundTrip(t *testing.T) {
	t.Parallel()

	encrypter, err := NewEncrypter("test-app-secret")
	if err != nil {
		t.Fatalf("NewEncrypter() error = %v", err)
	}

	plaintext := []byte("super-secret-token")
	ciphertext, nonce, err := encrypter.Encrypt(context.Background(), plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	if bytes.Equal(ciphertext, plaintext) {
		t.Fatal("ciphertext must differ from plaintext")
	}
	if len(nonce) == 0 {
		t.Fatal("nonce must not be empty")
	}

	decrypted, err := encrypter.Decrypt(context.Background(), ciphertext, nonce)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Fatalf("Decrypt() = %q, want %q", decrypted, plaintext)
	}
}

func TestEncrypter_InvalidNonce(t *testing.T) {
	t.Parallel()

	encrypter, err := NewEncrypter("test-app-secret")
	if err != nil {
		t.Fatalf("NewEncrypter() error = %v", err)
	}

	_, err = encrypter.Decrypt(context.Background(), []byte("cipher"), []byte("short"))
	if err == nil {
		t.Fatal("expected invalid nonce error")
	}
}
