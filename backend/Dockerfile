FROM golang:1.26-alpine AS base
WORKDIR /app

COPY backend/go.mod backend/go.sum ./backend/
WORKDIR /app/backend
RUN go mod download

WORKDIR /app
COPY api ./api
COPY backend ./backend

WORKDIR /app/backend
RUN cd internal/adapters/primary/http/openapi \
    && go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.7.0 \
        --config cfg.yaml ../../../../../../api/openapi.yaml

EXPOSE 8080
CMD ["go", "run", "./cmd/server"]
