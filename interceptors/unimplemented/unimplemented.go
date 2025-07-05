package unimplemented

import (
	"context"

	"connectrpc.com/connect"
)

// UnimplementedInterceptor in intended to be embedded in all interceptor structs to avoid filling out interface methods
// for concepts that dont make sense for that interceptor. It functions as a noop.
type UnimplementedInterceptor struct{}

func (u *UnimplementedInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		return next(ctx, req)
	})
}

func (u *UnimplementedInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		return next(ctx, conn)
	})
}

func (u *UnimplementedInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return connect.StreamingClientFunc(func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		return next(ctx, spec)
	})
}

