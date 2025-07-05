package server

import (
	"context"
	"fmt"
	"runtime"

	"connectrpc.com/connect"
	"github.com/go-logr/logr"
	"github.com/nicjohnson145/connecthelp/interceptors/unimplemented"
)

const (
	DefaultPanicStackBufferSize = 8192
)

type PanicInterceptorConfig struct {
	// Logger is the optional logger the panics will be logged with, if not given, will attempt to use the context
	// logger. If neither are present, no logging will be done
	Logger *logr.Logger
	// StackBufferSize is the optional stack size configuration. If not given will default to DefaultPanicStackBufferSize
	StackBufferSize *int
}

func NewPanicInterceptor(config PanicInterceptorConfig) *PanicInterceptor {
	interceptor := &PanicInterceptor{
		logger: config.Logger,
	}

	size := config.StackBufferSize
	if size == nil {
		interceptor.stackBufferSize = DefaultPanicStackBufferSize
	} else {
		interceptor.stackBufferSize = *size
	}

	return interceptor
}

var _ connect.Interceptor = (*PanicInterceptor)(nil)

type PanicInterceptor struct {
	unimplemented.UnimplementedInterceptor
	logger          *logr.Logger
	stackBufferSize int
}

func (p *PanicInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (resp connect.AnyResponse, err error) {
		log := p.getLogger(ctx)

		defer func() {
			if r := recover(); r != nil {
				stack := make([]byte, p.stackBufferSize)
				stack = stack[:runtime.Stack(stack, false)]
				log.Error(r.(error), "recovering from panic", "stack", string(stack))
				err = connect.NewError(connect.CodeInternal, fmt.Errorf("recovering from panic: %v", r))
			}
		}()

		resp, err = next(ctx, req)
		return resp, err
	})
}

func (p *PanicInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(ctx context.Context, conn connect.StreamingHandlerConn) (err error) {
		log := p.getLogger(ctx)

		defer func() {
			if r := recover(); r != nil {
				stack := make([]byte, p.stackBufferSize)
				stack = stack[:runtime.Stack(stack, false)]
				log.Error(r.(error), "recovering from panic", "stack", string(stack))
				err = connect.NewError(connect.CodeInternal, fmt.Errorf("recovering from panic: %v", r))
			}
		}()

		err = next(ctx, conn)
		return err
	})
}

func (p *PanicInterceptor) getLogger(ctx context.Context) logr.Logger {
	if p.logger != nil {
		return *p.logger
	} else {
		return logr.FromContextOrDiscard(ctx)
	}
}
