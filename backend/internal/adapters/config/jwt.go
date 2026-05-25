package config

import "os"

type JWT struct {
	Secret string
}

func LoadJWT() JWT {
	return JWT{Secret: os.Getenv("SESSION_SECRET")}
}
