package server

import (
	"fmt"
	"net/http"

	"github.com/nlypage/applycation/backend/internal/adapters/config"
	httpHandler "github.com/nlypage/applycation/backend/internal/adapters/primary/http/handler"
	"github.com/nlypage/applycation/backend/internal/adapters/primary/http/openapi"
	"github.com/swaggest/swgui/v5emb"
)

func NewHTTPServer(cfg config.HTTP, h *httpHandler.Handler) *http.Server {
	strict := openapi.NewStrictHandler(h, nil)
	mux := http.NewServeMux()

	mux.HandleFunc("GET /docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/", http.StatusTemporaryRedirect)
	})

	mux.Handle("GET /docs/", v5emb.New("applycation API", "/openapi.json", "/docs/"))

	mux.HandleFunc("GET /openapi.json", func(w http.ResponseWriter, _ *http.Request) {
		spec, err := openapi.GetSpecJSON()
		if err != nil {
			http.Error(w, "failed to load OpenAPI spec", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write(spec)
	})

	apiHandler := openapi.HandlerWithOptions(strict, openapi.StdHTTPServerOptions{BaseRouter: mux})

	return &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: apiHandler,
	}
}
