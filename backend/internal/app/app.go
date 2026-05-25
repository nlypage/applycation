package app

import (
	"log/slog"
	"net/http"

	"github.com/nlypage/applycation/backend/internal/adapters/config"
	httpHandler "github.com/nlypage/applycation/backend/internal/adapters/primary/http/handler"
	httpServer "github.com/nlypage/applycation/backend/internal/adapters/primary/http/server"
)

type App struct {
	HTTPServer *http.Server
}

func New() *App {
	provider := NewServiceProvider()
	handler := httpHandler.New(provider.HealthService)
	httpCfg := config.LoadHTTP()

	return &App{
		HTTPServer: httpServer.NewHTTPServer(httpCfg, handler),
	}
}

func (a *App) Run() error {
	slog.Info("starting backend server", "addr", a.HTTPServer.Addr)
	return a.HTTPServer.ListenAndServe()
}
