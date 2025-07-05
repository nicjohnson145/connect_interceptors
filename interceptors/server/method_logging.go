package server

import (
	"context"

	"connectrpc.com/connect"
	"github.com/go-logr/logr"
	"github.com/nicjohnson145/connecthelp/interceptors/unimplemented"
)

type MethodLoggingInterceptorConfig struct {
	// Logger is the optional logger the method calls will be logged with, if not given, will attempt to use the context
	// logger. If neither are present, no logging will be done
	Logger *logr.Logger
	// LogSuccessfulCompletion optionally will log a message on successful request completion
	LogSuccessfulCompletion bool
	// LogErrorCompletion optionally will log a message on request error
	LogErrorCompletion bool
}

func NewMethodLoggingInterceptor(config MethodLoggingInterceptorConfig) *MethodLoggingInterceptor {
	interceptor := &MethodLoggingInterceptor{
		logger:                  config.Logger,
		logSuccessfulCompletion: config.LogSuccessfulCompletion,
		logErrorCompletion:      config.LogErrorCompletion,
	}

	return interceptor
}

var _ connect.Interceptor = (*MethodLoggingInterceptor)(nil)

type MethodLoggingInterceptor struct {
	unimplemented.UnimplementedInterceptor
	logger                  *logr.Logger
	logSuccessfulCompletion bool
	logErrorCompletion      bool
}

func (m *MethodLoggingInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		log := m.getLogger(ctx).WithValues("path", req.Spec().Procedure)

		log.Info("request recieved")

		resp, err := next(ctx, req)

		if err != nil && m.logErrorCompletion {
			log.Error(err, "request completed with error")
		}
		if err == nil && m.logSuccessfulCompletion {
			log.Info("request completed")
		}

		return resp, err
	})
}

func (m *MethodLoggingInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		log := m.getLogger(ctx).WithValues("path", conn.Spec().Procedure)

		log.Info("stream started")

		err := next(ctx, conn)

		if err != nil && m.logErrorCompletion {
			log.Error(err, "stream ended with error")
		}
		if err == nil && m.logSuccessfulCompletion {
			log.Info("stream completed")
		}

		return err
	})
}

func (m *MethodLoggingInterceptor) getLogger(ctx context.Context) logr.Logger {
	if m.logger != nil {
		return *m.logger
	} else {
		return logr.FromContextOrDiscard(ctx)
	}
}
