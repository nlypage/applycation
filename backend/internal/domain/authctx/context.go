package authctx

import "context"

type contextKey string

const ownerIDKey contextKey = "owner_id"

func WithOwnerID(ctx context.Context, ownerID string) context.Context {
	return context.WithValue(ctx, ownerIDKey, ownerID)
}

func OwnerID(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ownerIDKey).(string)
	return v, ok
}
