package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/nlypage/applycation/backend/internal/adapters/config"
	httpHandler "github.com/nlypage/applycation/backend/internal/adapters/primary/http/handler"
	httpServer "github.com/nlypage/applycation/backend/internal/adapters/primary/http/server"
	"github.com/nlypage/applycation/backend/pkg/closer"
)

const httpShutdownTimeout = 5 * time.Second

type App struct {
	HTTPServer *http.Server

	provider *ServiceProvider
}

func New() *App {
	closeManager := closer.New(os.Interrupt, syscall.SIGTERM)
	provider := NewServiceProvider(closeManager)
	handler := httpHandler.New(provider.HealthService)
	httpCfg := config.LoadHTTP()
	httpSrv := httpServer.NewHTTPServer(httpCfg, handler)

	provider.AddCloser(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), httpShutdownTimeout)
		defer cancel()

		slog.Info("shutting down backend server", "addr", httpSrv.Addr)
		if err := httpSrv.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("shutdown http server: %w", err)
		}

		return nil
	})

	return &App{
		HTTPServer: httpSrv,
		provider:   provider,
	}
}

func (a *App) Run() error {
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("starting backend server", "addr", a.HTTPServer.Addr)
		if err := a.HTTPServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	closed := make(chan struct{})
	go func() {
		a.provider.Wait()
		close(closed)
	}()

	select {
	case err := <-serverErr:
		a.provider.CloseAll()
		<-closed
		return err
	case <-closed:
		return <-serverErr
	}
}
