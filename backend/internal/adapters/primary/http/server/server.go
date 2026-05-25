package server

import (
	"fmt"
	"net/http"

	"github.com/nlypage/applycation/backend/internal/adapters/config"
	httpHandler "github.com/nlypage/applycation/backend/internal/adapters/primary/http/handler"
	"github.com/nlypage/applycation/backend/internal/adapters/primary/http/openapi"
)

func NewHTTPServer(cfg config.HTTP, h *httpHandler.Handler) *http.Server {
	strict := openapi.NewStrictHandler(h, nil)
	apiHandler := openapi.Handler(strict)

	return &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: apiHandler,
	}
}
