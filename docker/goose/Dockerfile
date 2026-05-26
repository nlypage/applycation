FROM golang:1.26-alpine

RUN go install github.com/pressly/goose/v3/cmd/goose@v3.27.1

ENTRYPOINT ["/go/bin/goose"]
