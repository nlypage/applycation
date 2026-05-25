package config

import "os"

const defaultPort = "8080"

type HTTP struct {
	Port string
}

func LoadHTTP() HTTP {
	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		port = defaultPort
	}

	return HTTP{Port: port}
}
