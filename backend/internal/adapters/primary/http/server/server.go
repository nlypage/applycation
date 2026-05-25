package server

import (
	"fmt"
	"net/http"

	"github.com/nlypage/applycation/backend/internal/adapters/config"
	httpHandler "github.com/nlypage/applycation/backend/internal/adapters/primary/http/handler"
	"github.com/nlypage/applycation/backend/internal/adapters/primary/http/openapi"
)

const redocHTML = `<!doctype html>
<html>
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width,initial-scale=1" />
    <title>applycation API docs</title>
    <style>
      body { margin: 0; }
    </style>
  </head>
  <body>
    <redoc spec-url="/openapi.json"></redoc>
    <script src="https://cdn.jsdelivr.net/npm/redoc@next/bundles/redoc.standalone.js"></script>
  </body>
</html>`

func NewHTTPServer(cfg config.HTTP, h *httpHandler.Handler) *http.Server {
	strict := openapi.NewStrictHandler(h, nil)
	mux := http.NewServeMux()

	mux.HandleFunc("GET /docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/", http.StatusTemporaryRedirect)
	})

	mux.HandleFunc("GET /docs/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(redocHTML))
	})

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
