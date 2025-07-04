package server

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/go-logr/logr"
	"github.com/nicjohnson145/connect_interceptors/unimplemented"
	"github.com/nicjohnson145/hlp/set"
)

var (
	ErrIncludedExcludedMutallyExclusiveError = errors.New("IncludedMethods and ExcludedMethods are mutually exclusive")
)

const (
	DefaultSlowdownAmount = 3 * time.Second
)

type SlowdownInterceptorConfig struct {
	// Logger is the optional logger the method calls will be logged with, if not given, will attempt to use the context
	// logger. If neither are present, no logging will be done
	Logger *logr.Logger
	// Amount is the optional amount to slow down responses by, if not given will default to DefaultSlowdownAmount
	Amount *time.Duration
	// IncludedMethods is the comma-separated list of methods that should be slowed down. The special value '*' means
	// slow down all methods. This configuration param is mutally exclusive with ExcludedMethods
	IncludedMethods string
	// ExcludedMethods is the comma-separated list of mehtods that should NOT be slowed down. This configuration param
	// is mutally exclusive with IncludedMethods
	ExcludedMethods string
}

func NewSlowdownInterceptor(config SlowdownInterceptorConfig) (*SlowdownInterceptor, error) {
	if config.ExcludedMethods != "" && config.IncludedMethods != "" {
		return nil, ErrIncludedExcludedMutallyExclusiveError
	}

	toFilter := func(inclusion string, exclusion string) slowdownFilter {
		if inclusion != "" {
			switch inclusion {
			case "*":
				return func(str string) bool { return true }
			default:
				methodSet := set.New(strings.Split(inclusion, ",")...)
				return func(str string) bool { return methodSet.Contains(str) }
			}
		} else if exclusion != "" {
			methodSet := set.New(strings.Split(exclusion, ",")...)
			return func(str string) bool { return !methodSet.Contains(str) }
		} else {
			return func(str string) bool { return false }
		}
	}

	amount := DefaultSlowdownAmount
	if config.Amount != nil {
		amount = *config.Amount
	}

	return &SlowdownInterceptor{
		logger: config.Logger,
		amount: amount,
		filter: toFilter(config.IncludedMethods, config.ExcludedMethods),
	}, nil
}

var _ connect.Interceptor = (*SlowdownInterceptor)(nil)

type slowdownFilter func(str string) bool

// SlowdownInterceptor slows down responses, mostly useful in development testing where you want to test loading
// states/etc but dont want to use browser network throttling so _everything_ is slow.
type SlowdownInterceptor struct {
	unimplemented.UnimplementedInterceptor
	logger *logr.Logger
	filter slowdownFilter
	amount time.Duration
}

func (p *SlowdownInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		resp, err := next(ctx, req)

		if p.filter(req.Spec().Procedure) {
			log := p.getLogger(ctx)
			log.V(1).Info(fmt.Sprintf("slowing down response %v", p.amount), "interceptor", "slowdown")
			time.Sleep(p.amount)
		}

		return resp, err
	})
}

// TODO: implement streaming. Probably implmenting the StreamingHandlerConn interface, and "peeking" the initial client
// request, and then logging the responses in the overwritten `Send` method
func (p *SlowdownInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		return next(ctx, conn)
	})
}

func (p *SlowdownInterceptor) getLogger(ctx context.Context) logr.Logger {
	if p.logger != nil {
		return *p.logger
	} else {
		return logr.FromContextOrDiscard(ctx)
	}
}
