package handler

import (
	"context"
	"testing"

	"github.com/nlypage/applycation/backend/internal/adapters/primary/http/openapi"
	"github.com/nlypage/applycation/backend/internal/domain/entity"
	primarymocks "github.com/nlypage/applycation/backend/internal/testutil/mocks/primary"
)

func TestHandlerGetHealth(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	healthService := primarymocks.NewMockHealthService(t)
	healthService.EXPECT().Health(ctx).Return(entity.Health{Status: "ok"}).Once()

	response, err := New(healthService).GetHealth(ctx, openapi.GetHealthRequestObject{})
	if err != nil {
		t.Fatalf("GetHealth() unexpected error: %v", err)
	}

	got, ok := response.(openapi.GetHealth200JSONResponse)
	if !ok {
		t.Fatalf("GetHealth() response type = %T, want %T", response, openapi.GetHealth200JSONResponse{})
	}
	if got.Status != "ok" {
		t.Fatalf("GetHealth().Status = %q, want %q", got.Status, "ok")
	}
}
