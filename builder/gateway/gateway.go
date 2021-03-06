package gateway

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/bluekaki/vv/internal/configs"
	"github.com/bluekaki/vv/internal/interceptor"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/resolver/dns"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	defaultKeepAlive = &keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             time.Second,
		PermitWithoutStream: true,
	}

	defaultDialTimeout = time.Second * 2
)

func init() {
	runtime.DefaultContextTimeout = time.Second * 10
}

// Option how setup client
type Option func(*option)

type option struct {
	credential  credentials.TransportCredentials
	keepalive   *keepalive.ClientParameters
	dialTimeout time.Duration
}

// WithCredential setup credential for tls
func WithCredential(credential credentials.TransportCredentials) Option {
	return func(opt *option) {
		opt.credential = credential
	}
}

// WithKeepAlive setup keepalive parameters
func WithKeepAlive(keepalive *keepalive.ClientParameters) Option {
	return func(opt *option) {
		opt.keepalive = keepalive
	}
}

// WithDialTimeout setup the dial timeout
func WithDialTimeout(timeout time.Duration) Option {
	return func(opt *option) {
		opt.dialTimeout = timeout
	}
}

// New create grpc-gateway server mux, and grpc dial options.
func New(options ...Option) (*runtime.ServeMux, []grpc.DialOption) {
	opt := new(option)
	for _, f := range options {
		f(opt)
	}

	kacp := defaultKeepAlive
	if opt.keepalive != nil {
		kacp = opt.keepalive
	}

	dialTimeout := defaultDialTimeout
	if opt.dialTimeout > 0 {
		dialTimeout = opt.dialTimeout
	}

	mux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(runtime.DefaultHeaderMatcher),
		runtime.WithOutgoingHeaderMatcher(runtime.DefaultHeaderMatcher),
		runtime.WithMetadata(annotator),
		runtime.WithErrorHandler(runtime.DefaultHTTPErrorHandler),
		runtime.WithStreamErrorHandler(runtime.DefaultStreamErrorHandler),
		runtime.WithRoutingErrorHandler(runtime.DefaultRoutingErrorHandler),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.HTTPBodyMarshaler{
			Marshaler: &runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					UseProtoNames:   true,
					EmitUnpopulated: true,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					DiscardUnknown: true,
				},
			},
		}),
	)

	gatewayInterceptor := interceptor.NewGatewayInterceptor()

	dialOptions := []grpc.DialOption{
		grpc.WithResolvers(dns.NewBuilder()),
		grpc.WithTimeout(dialTimeout),
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(*kacp),
		grpc.WithUnaryInterceptor(gatewayInterceptor.UnaryInterceptor),
		grpc.WithDefaultServiceConfig(configs.ServiceConfig),
	}

	if opt.credential == nil {
		dialOptions = append(dialOptions, grpc.WithInsecure())
	} else {
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(opt.credential))
	}

	return mux, dialOptions
}

func annotator(ctx context.Context, req *http.Request) metadata.MD {
	body, _ := ioutil.ReadAll(req.Body)
	req.Body = ioutil.NopCloser(bytes.NewBuffer(body)) // re-construct req body

	return metadata.Pairs(
		interceptor.Authorization, req.Header.Get("Authorization"),
		interceptor.ProxyAuthorization, req.Header.Get("Proxy-Authorization"),
		interceptor.Date, req.Header.Get("Date"),
		interceptor.Method, req.Method,
		interceptor.URI, req.RequestURI,
		interceptor.Body, string(body), // TODO unsafe
		interceptor.XForwardedFor, req.Header.Get("X-Forwarded-For"),
		interceptor.XForwardedHost, req.Header.Get("X-Forwarded-Host"),
	)
}
