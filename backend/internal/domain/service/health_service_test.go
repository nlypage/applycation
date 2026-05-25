package service

import (
	"context"
	"testing"
)

func TestHealthServiceHealth(t *testing.T) {
	t.Parallel()

	svc := NewHealthService()

	got := svc.Health(context.Background())
	if got.Status != "ok" {
		t.Fatalf("Health().Status = %q, want %q", got.Status, "ok")
	}
}
