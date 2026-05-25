package config

import "os"

type Postgres struct {
	DatabaseURL string
}

func LoadPostgres() Postgres {
	return Postgres{
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
}
