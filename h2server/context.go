package h2server

import "context"

type connKey struct{}

func contextWithConn(ctx context.Context, conn *pseudoConn) context.Context {
	return context.WithValue(ctx, connKey{}, conn)
}

func connFromContext(ctx context.Context) *pseudoConn {
	return ctx.Value(connKey{}).(*pseudoConn)
}
