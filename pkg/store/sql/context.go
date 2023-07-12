package sql

import (
	"context"
)

type contextKey string

const querierContextKey contextKey = "querier"

func NewContext(ctx context.Context, querier Querier) context.Context {
	return context.WithValue(ctx, querierContextKey, querier)
}

func FromContext(ctx context.Context, original Querier) Querier {
	querier, ok := ctx.Value(querierContextKey).(Querier)
	if !ok {
		return original
	}
	return querier
}
