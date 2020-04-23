package util

import (
	"context"
	"time"

	"google.golang.org/grpc/metadata"
)

// CtxSubDeadline sub ctx deadline for relay context
func CtxSubDeadline(ctx context.Context, sub time.Duration,
	defaultTimeout time.Duration) (context.Context, context.CancelFunc) {

	if deadline, ok := ctx.Deadline(); ok {
		return context.WithDeadline(ctx, deadline.Add(-1*sub))
	}

	if defaultTimeout != 0 {
		return context.WithTimeout(ctx, defaultTimeout)
	}

	return ctx, context.CancelFunc(func() {})
}

// CtxRelayMD relay context Value for grpc
func CtxRelayMD(ctx context.Context) context.Context {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		ctx = metadata.NewOutgoingContext(ctx, md)
	}
	return ctx
}
