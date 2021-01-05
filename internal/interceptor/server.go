package interceptor

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"runtime/debug"
	"strings"
	"time"

	"github.com/bluekaki/vv/internal/protos/gen"
	"github.com/bluekaki/vv/options"

	protoV1 "github.com/golang/protobuf/proto"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/koketama/minami58"
	"github.com/koketama/pbutil"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	// JournalID a random id used by log journal
	JournalID = "journal_id"
	// Authorization used by auth, both gateway and grpc
	Authorization = "authorization"
	// ProxyAuthorization used by signature, both gateway and grpc
	ProxyAuthorization = "proxy-authorization"
	// Date GMT format
	Date = "date"
	// Method http.XXXMethod
	Method = "method"
	// URI url encoded
	URI = "uri"
	// Body string body
	Body = "body"
	// XForwardedFor forwarded for
	XForwardedFor = "x-forwarded-for"
	// XForwardedHost forwarded host
	XForwardedHost = "x-forwarded-host"
)

// SessionUserinfo mark userinfo in context
type SessionUserinfo struct{}

var toLoggedMetadata = map[string]bool{
	Authorization:      true,
	ProxyAuthorization: true,
	Date:               true,
	Method:             true,
	URI:                true,
	Body:               true,
	XForwardedFor:      true,
	XForwardedHost:     true,
}

var _ Payload = (*restPayload)(nil)
var _ Payload = (*grpcPayload)(nil)

// Payload rest or grpc payload
type Payload interface {
	JournalID() string
	ForwardedByGrpcGateway() bool
	Service() string
	Date() string
	Method() string
	URI() string
	Body() string
	t()
}

type restPayload struct {
	journalID string
	service   string
	date      string
	method    string
	uri       string
	body      string
}

func (r *restPayload) JournalID() string {
	return r.journalID
}

func (r *restPayload) ForwardedByGrpcGateway() bool {
	return true
}

func (r *restPayload) Service() string {
	return r.service
}

func (r *restPayload) Date() string {
	return r.date
}

func (r *restPayload) Method() string {
	return r.method
}

func (r *restPayload) URI() string {
	return r.uri
}

func (r *restPayload) Body() string {
	return r.body
}

func (r *restPayload) t() {}

type grpcPayload struct {
	journalID string
	service   string
	date      string
	method    string
	uri       string
	body      string
}

func (g *grpcPayload) JournalID() string {
	return g.journalID
}

func (g *grpcPayload) ForwardedByGrpcGateway() bool {
	return false
}

func (g *grpcPayload) Service() string {
	return g.service
}

func (g *grpcPayload) Date() string {
	return g.date
}

func (g *grpcPayload) Method() string {
	return g.method
}

func (g *grpcPayload) URI() string {
	return g.uri
}

func (g *grpcPayload) Body() string {
	return g.body
}

func (g *grpcPayload) t() {}

// NewServerInterceptor create a server interceptor
func NewServerInterceptor(logger *zap.Logger) *ServerInterceptor {
	return &ServerInterceptor{
		logger: logger,
	}
}

// ServerInterceptor the server's interceptor
type ServerInterceptor struct {
	logger *zap.Logger
}

func (s *ServerInterceptor) journalID() string {
	nonce := make([]byte, 16)
	io.ReadFull(rand.Reader, nonce)

	return string(minami58.Encode(nonce))
}

// UnaryInterceptor a interceptor for server unary operations
func (s *ServerInterceptor) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	ts := time.Now()
	journalID := s.journalID()

	fullMethod := strings.Split(info.FullMethod, "/")
	serviceName := fullMethod[1]
	methodName := fullMethod[2]

	doJournal := false
	if proto.GetExtension(FileDescriptor.Options(info.FullMethod), options.E_Journal).(bool) {
		doJournal = true
	}

	defer func() { // double recover for safety
		if p := recover(); p != nil {
			s, _ := status.New(codes.Internal, fmt.Sprintf("got double panic => journal_id: %s, error: %+v", journalID, p)).WithDetails(&pb.Stack{Info: string(debug.Stack())})
			err = s.Err()
		}
	}()

	defer func() {
		if p := recover(); p != nil {
			s, _ := status.New(codes.Internal, fmt.Sprintf("got panic => journal_id: %s, error: %+v", journalID, p)).WithDetails(&pb.Stack{Info: string(debug.Stack())})
			err = s.Err()
		}

		grpc.SetHeader(ctx, metadata.Pairs(runtime.MetadataHeaderPrefix+JournalID, journalID))

		if doJournal {
			journal := &pb.Journal{
				Id: journalID,
				Request: &pb.Request{
					Restapi: ForwardedByGrpcGateway(ctx),
					Method:  info.FullMethod,
					Metadata: func() map[string]string {
						meta, _ := metadata.FromIncomingContext(ctx)
						mp := make(map[string]string)
						for key, values := range meta {
							if toLoggedMetadata[key] {
								mp[key] = values[0]
							}
						}
						return mp
					}(),
					Payload: func() *anypb.Any {
						if req == nil {
							return nil
						}

						any, _ := anypb.New(req.(proto.Message))
						return any
					}(),
				},
				Response: &pb.Response{
					Code: codes.OK.String(),
					Payload: func() *anypb.Any {
						if resp == nil {
							return nil
						}

						any, _ := anypb.New(resp.(proto.Message))
						return any
					}(),
				},
				Success: err == nil,
			}

			if err != nil {
				s, _ := status.FromError(err)
				journal.Response.Code = s.Code().String()
				journal.Response.Message = s.Message()

				journal.Response.Details = make([]*anypb.Any, len(s.Details()))
				for i, detail := range s.Details() {
					journal.Response.Details[i], _ = anypb.New(detail.(proto.Message))
				}
			}

			journal.CostSeconds = time.Since(ts).Seconds()

			json, _ := pbutil.ProtoMessage2Map(journal)
			if err == nil {
				s.logger.Info("unary interceptor", zap.Any("journal", json))
			} else {
				s.logger.Error("unary interceptor", zap.Any("journal", json))
			}
		}

		method := info.FullMethod

		if http := proto.GetExtension(FileDescriptor.Options(info.FullMethod), annotations.E_Http).(*annotations.HttpRule); http != nil {
			if x, ok := http.GetPattern().(*annotations.HttpRule_Get); ok {
				method = "get " + x.Get
			} else if x, ok := http.GetPattern().(*annotations.HttpRule_Put); ok {
				method = "put " + x.Put
			} else if x, ok := http.GetPattern().(*annotations.HttpRule_Post); ok {
				method = "post " + x.Post
			} else if x, ok := http.GetPattern().(*annotations.HttpRule_Delete); ok {
				method = "delete " + x.Delete
			} else if x, ok := http.GetPattern().(*annotations.HttpRule_Patch); ok {
				method = "patch " + x.Patch
			}
		}

		if alias := proto.GetExtension(FileDescriptor.Options(info.FullMethod), options.E_MetricsAlias).(string); alias != "" {
			method = alias
		}

		if err == nil {
			MetricsRequestCost.WithLabelValues(method).Observe(time.Since(ts).Seconds())
		} else {
			MetricsError.WithLabelValues(method, status.Code(err).String(), err.Error(), journalID).Observe(time.Since(ts).Seconds())
		}
	}()

	meta, _ := metadata.FromIncomingContext(ctx)
	meta.Set(JournalID, journalID)
	ctx = metadata.NewOutgoingContext(ctx, meta)

	var (
		authorizationValidator      userinfoHandler
		proxyAuthorizationValidator signatureHandler
	)
	if option := proto.GetExtension(FileDescriptor.Options(info.FullMethod), options.E_Authorization).(*options.Validator); option != nil {
		authorizationValidator = Validator.AuthorizationValidator(option.Name)
	}
	if option := proto.GetExtension(FileDescriptor.Options(info.FullMethod), options.E_ProxyAuthorization).(*options.Validator); option != nil {
		proxyAuthorizationValidator = Validator.ProxyAuthorizationValidator(option.Name)
	}

	if authorizationValidator == nil && proxyAuthorizationValidator == nil {
		return handler(ctx, req)
	}

	var auth, proxyAuth string
	if authHeader := meta.Get(Authorization); len(authHeader) != 0 {
		auth = authHeader[0]
	}
	if proxyAuthHeader := meta.Get(ProxyAuthorization); len(proxyAuthHeader) != 0 {
		proxyAuth = proxyAuthHeader[0]
	}

	var payload Payload
	if forwardedByGrpcGateway(meta) {
		payload = &restPayload{
			journalID: journalID,
			service:   serviceName,
			date:      meta.Get(Date)[0],
			method:    meta.Get(Method)[0],
			uri:       meta.Get(URI)[0],
			body:      meta.Get(Body)[0],
		}

	} else {
		payload = &grpcPayload{
			journalID: journalID,
			service:   serviceName,
			date:      meta.Get(Date)[0],
			method:    methodName,
			uri:       info.FullMethod,
			body: func() string {
				if req == nil {
					return ""
				}

				raw, _ := pbutil.ProtoMessage2JSON(req.(protoV1.Message))
				return raw
			}(),
		}
	}

	if authorizationValidator != nil {
		userinfo, err := authorizationValidator(auth, payload)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, fmt.Sprintf("%+v", err))
		}
		ctx = context.WithValue(ctx, SessionUserinfo{}, userinfo)
	}

	if proxyAuthorizationValidator != nil {
		ok, err := proxyAuthorizationValidator(proxyAuth, payload)
		if err != nil {
			return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("%+v", err))
		}
		if !ok {
			return nil, status.Error(codes.PermissionDenied, codes.PermissionDenied.String())
		}
	}

	return handler(ctx, req)
}

type serverWrappedStream struct {
	grpc.ServerStream
}

func (s *serverWrappedStream) RecvMsg(m interface{}) (err error) {
	return s.ServerStream.RecvMsg(m)
}

func (s *serverWrappedStream) SendMsg(m interface{}) (err error) {
	return s.ServerStream.SendMsg(m)
}

// StreamInterceptor a interceptor for server stream operations
func (s *ServerInterceptor) StreamInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	// TODO
	return errors.New("Not currently supported")

	// return handler(srv, &serverWrappedStream{ServerStream: stream})
}
