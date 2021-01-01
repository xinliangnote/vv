package interceptor

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/bluekaki/vv/internal/protos/gen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var gwHeader = struct {
	key   string
	value string
}{
	key:   "grpc-gateway",
	value: "koketama/grpcgw/m/v1.0",
}

// ForwardedByGrpcGateway whether forwarded by grpc gateway
func ForwardedByGrpcGateway(ctx context.Context) bool {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false
	}

	return forwardedByGrpcGateway(meta)
}

func forwardedByGrpcGateway(meta metadata.MD) bool {
	values := meta.Get(gwHeader.key)
	if len(values) == 0 {
		return false
	}

	return values[0] == gwHeader.value
}

// NewGatewayInterceptor create a gateway interceptor
func NewGatewayInterceptor() *GatewayInterceptor {
	return new(GatewayInterceptor)
}

// GatewayInterceptor the gateway's interceptor
type GatewayInterceptor struct {
}

// UnaryInterceptor a interceptor for gateway unary operations
func (g *GatewayInterceptor) UnaryInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	defer func() {
		if p := recover(); p != nil {
			s, _ := status.New(codes.Internal, fmt.Sprintf("%+v", p)).WithDetails(&pb.Stack{Info: string(debug.Stack())})
			err = s.Err()
		}
	}()

	meta, _ := metadata.FromOutgoingContext(ctx)
	if meta == nil {
		meta = make(metadata.MD)
	}

	// TODO verify auth in future

	meta.Set(gwHeader.key, gwHeader.value)
	ctx = metadata.NewOutgoingContext(ctx, meta)

	return invoker(ctx, method, req, reply, cc, opts...)
}
