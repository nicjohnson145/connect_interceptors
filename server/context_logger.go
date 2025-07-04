package server

import (
	"context"

	"connectrpc.com/connect"
	"github.com/go-logr/logr"
	"github.com/nicjohnson145/connect_interceptors/unimplemented"
	"github.com/oklog/ulid/v2"
)

type ContextLoggerInterceptorConfig struct {
	// RootLogger is the logger that all request level loggers will inherit from
	RootLogger logr.Logger
	// NoAttachRequestID indicates that a request ulid should not be generated and attached to the logger
	NoAttachRequestID bool
}

func NewContextLoggerInterceptor(config ContextLoggerInterceptorConfig) *ContextLoggerInterceptor {
	return &ContextLoggerInterceptor{
		rootLogger:      config.RootLogger,
		attachRequestID: !config.NoAttachRequestID,
	}
}

var _ connect.Interceptor = (*ContextLoggerInterceptor)(nil)

type ContextLoggerInterceptor struct {
	unimplemented.UnimplementedInterceptor
	rootLogger      logr.Logger
	attachRequestID bool
}

func (c *ContextLoggerInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		return next(c.embed(ctx), req)
	})
}

func (c *ContextLoggerInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		return next(c.embed(ctx), conn)
	})
}

func (c *ContextLoggerInterceptor) embed(ctx context.Context) context.Context {
	reqLogger := c.rootLogger

	if c.attachRequestID {
		reqLogger = c.rootLogger.WithValues("request-id", ulid.Make().String())
	}

	return logr.NewContext(ctx, reqLogger)
}
