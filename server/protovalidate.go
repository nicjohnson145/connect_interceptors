package server

import (
	"context"
	"strings"

	"connectrpc.com/connect"
	"buf.build/go/protovalidate"
	"github.com/nicjohnson145/connect_interceptors/unimplemented"
	"github.com/nicjohnson145/hlp/set"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type ProtovalidateInterceptorConfig struct {
	// SkipMethods is a comma separated list of methods to skip validation for
	SkipMethods string
}

func NewProtovalidateInterceptor(config ProtovalidateInterceptorConfig) *ProtovalidateInterceptor {
	toFilter := func(str string) skipFilter {
		switch str {
		case "":
			return func(s string) bool { return false }
		default:
			methodSet := set.New(strings.Split(str, ",")...)
			return func(s string) bool { return methodSet.Contains(s) }
		}
	}

	return &ProtovalidateInterceptor{
		skipFilter: toFilter(config.SkipMethods),
	}
}

var _ connect.Interceptor = (*ProtovalidateInterceptor)(nil)

type skipFilter func(str string) bool

type ProtovalidateInterceptor struct {
	unimplemented.UnimplementedInterceptor
	skipFilter skipFilter
}
func (p *ProtovalidateInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if !p.skipFilter(req.Spec().Procedure) {
			if objProto, ok := req.Any().(protoreflect.ProtoMessage); ok {
				if err := protovalidate.Validate(objProto); err != nil {
					return nil, connect.NewError(connect.CodeInvalidArgument, err)
				}
			}
		}
		return next(ctx, req)
	})
}

// TODO: implement streaming. Probably implmenting the StreamingHandlerConn interface, and "peeking" the initial client
// request, and then logging the responses in the overwritten `Send` method
func (p *ProtovalidateInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		return next(ctx, conn)
	})
}
