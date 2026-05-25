package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/nlypage/applycation/backend/internal/app"
)

func main() {
	application := app.New()
	if err := application.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
