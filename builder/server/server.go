package server

import (
	"time"

	"github.com/bluekaki/vv/internal/interceptor"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/keepalive"
)

var (
	defaultEnforcementPolicy = &keepalive.EnforcementPolicy{
		MinTime:             5 * time.Second,
		PermitWithoutStream: true,
	}

	defaultKeepAlive = &keepalive.ServerParameters{
		MaxConnectionIdle:     15 * time.Second,
		MaxConnectionAge:      30 * time.Second,
		MaxConnectionAgeGrace: 5 * time.Second,
		Time:                  5 * time.Second,
		Timeout:               1 * time.Second,
	}
)

// Option how setup client
type Option func(*option)

type option struct {
	credential        credentials.TransportCredentials
	enforcementPolicy *keepalive.EnforcementPolicy
	keepalive         *keepalive.ServerParameters
	prometheusHandler func(*zap.Logger)
}

// WithCredential setup credential for tls
func WithCredential(credential credentials.TransportCredentials) Option {
	return func(opt *option) {
		opt.credential = credential
	}
}

// WithEnforcementPolicy setup enforcement policy
func WithEnforcementPolicy(enforcementPolicy *keepalive.EnforcementPolicy) Option {
	return func(opt *option) {
		opt.enforcementPolicy = enforcementPolicy
	}
}

// WithKeepAlive setup keepalive parameters
func WithKeepAlive(keepalive *keepalive.ServerParameters) Option {
	return func(opt *option) {
		opt.keepalive = keepalive
	}
}

// New create a grpc server
func New(logger *zap.Logger, options ...Option) (*grpc.Server, error) {
	if logger == nil {
		return nil, errors.New("logger required")
	}

	opt := new(option)
	for _, f := range options {
		f(opt)
	}

	if opt.prometheusHandler != nil {
		opt.prometheusHandler(logger)
	}

	enforcementPolicy := defaultEnforcementPolicy
	if opt.enforcementPolicy != nil {
		enforcementPolicy = opt.enforcementPolicy
	}

	keepalive := defaultKeepAlive
	if opt.keepalive != nil {
		keepalive = opt.keepalive
	}

	serverInterceptor := interceptor.NewServerInterceptor(logger, opt.prometheusHandler != nil)

	serverOptions := []grpc.ServerOption{
		grpc.KeepaliveEnforcementPolicy(*enforcementPolicy),
		grpc.KeepaliveParams(*keepalive),
		grpc.UnaryInterceptor(serverInterceptor.UnaryInterceptor),
		grpc.StreamInterceptor(serverInterceptor.StreamInterceptor),
	}

	if opt.credential != nil {
		serverOptions = append(serverOptions, grpc.Creds(opt.credential))
	}

	return grpc.NewServer(serverOptions...), nil
}
