package server

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/go-logr/logr"
	"github.com/nicjohnson145/connect_interceptors/unimplemented"
	"github.com/nicjohnson145/hlp/set"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type PayloadLoggingInterceptorConfig struct {
	// Logger is the optional logger the method calls will be logged with, if not given, will attempt to use the context
	// logger. If neither are present, no logging will be done
	Logger *logr.Logger
	// RequestMethods is a comma separated list of methods to log requests for. The special value of '*' means log all
	// request payloads
	RequestMethods string
	// ResponseMethods is a comma separated list of methods to log responses for. The special value of '*' means log all
	// response payloads
	ResponseMethods string
	// Pretty eschew's the provided logger & context logger, and instead prints the output with fmt.Println ad a
	// human-readable indented object. Mostly intended for development debugging where log aggregators maybe not be in
	// play
	Pretty bool
}

func NewPayloadLoggingInterceptor(config PayloadLoggingInterceptorConfig) *PayloadLoggingInterceptor {
	toFilter := func(str string) methodFilter {
		switch str {
		case "":
			return func(s string) bool { return false }
		case "*":
			return func(s string) bool { return true }
		default:
			methodSet := set.New(strings.Split(str, ",")...)
			return func(s string) bool { return methodSet.Contains(s) }
		}
	}

	return &PayloadLoggingInterceptor{
		logger:         config.Logger,
		requestFilter:  toFilter(config.RequestMethods),
		responseFilter: toFilter(config.ResponseMethods),
		pretty:         config.Pretty,
	}
}

var _ connect.Interceptor = (*PayloadLoggingInterceptor)(nil)

type methodFilter func(string) bool

type PayloadLoggingInterceptor struct {
	unimplemented.UnimplementedInterceptor
	logger *logr.Logger

	requestFilter  methodFilter
	responseFilter methodFilter
	pretty         bool
}

func (p *PayloadLoggingInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		log := p.getLogger(ctx)

		maybeLog := func(filter methodFilter, method string, obj any, objType string) {
			if !filter(method) {
				return
			}

			objProto, ok := obj.(protoreflect.ProtoMessage)
			if ok {
				opts := protojson.MarshalOptions{}
				if p.pretty {
					opts.Indent = "    "
				}

				out, err := opts.Marshal(objProto)
				if err != nil {
					log.Error(err, fmt.Sprintf("unable to marshal object using protojson, cannot log %v", objType))
					return
				}

				if p.pretty {
					fmt.Println("\n" + string(out))
				} else {
					log.Info(objType, "object", string(out))
				}
			}
		}

		maybeLog(p.requestFilter, req.Spec().Procedure, req, "request object")
		resp, err := next(ctx, req)
		if resp != nil {
			maybeLog(p.responseFilter, req.Spec().Procedure, resp, "response object")
		}

		return resp, err
	})
}

// TODO: implement streaming. Probably implmenting the StreamingHandlerConn interface, and "peeking" the initial client
// request, and then logging the responses in the overwritten `Send` method
func (p *PayloadLoggingInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		return next(ctx, conn)
	})
}

func (p *PayloadLoggingInterceptor) getLogger(ctx context.Context) logr.Logger {
	if p.logger != nil {
		return *p.logger
	} else {
		return logr.FromContextOrDiscard(ctx)
	}
}
