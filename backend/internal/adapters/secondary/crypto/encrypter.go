package crypto

import "context"

type Encrypter struct{}

func NewEncrypter() *Encrypter {
	return &Encrypter{}
}

func (e *Encrypter) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	_ = ctx
	return plaintext, nil
}
