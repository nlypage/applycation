package secondary

import "context"

type Encrypter interface {
	Encrypt(ctx context.Context, plaintext []byte) (ciphertext []byte, nonce []byte, err error)
	Decrypt(ctx context.Context, ciphertext []byte, nonce []byte) ([]byte, error)
	Algorithm() string
}
