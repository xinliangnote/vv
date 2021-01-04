package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/bluekaki/vv"
	vvs "github.com/bluekaki/vv/builder/server"
	"github.com/bluekaki/vv/test/testdata/pb/gen"

	"github.com/koketama/auth"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func init() {
	vvs.RegisteAuthorizationValidator("userinfo_handler", func(authorization string, payload vvs.Payload) (userinfo interface{}, err error) {
		if authorization != "dummy token" && authorization != "minami" {
			return nil, errors.New("illegal authorization")
		}

		return &struct{ UserName string }{UserName: "minami"}, nil
	})

	vvs.RegisteProxyAuthorizationValidator("signature_handler", func(proxyAuthorization string, payload vvs.Payload) (ok bool, err error) {
		method := auth.MethodGRPC
		if payload.ForwardedByGrpcGateway() {
			method = auth.ToMethod(payload.Method())
		}

		_, ok, err = signature.Verify(proxyAuthorization, payload.Date(), method, payload.URI(), []byte(payload.Body()))
		return
	})
}

func newServer(grpcAddr, prometheusAddr string) *grpc.Server {
	server, err := vvs.New(logger, vvs.WithPrometheus(prometheusAddr))
	if err != nil {
		logger.Fatal("new server err", zap.Error(err))
	}

	pb.RegisterHelloServiceServer(server, new(helloServer), vvs.ParseFileDescriptorP)
	pb.RegisterDummyServiceServer(server, new(dummyServer), vvs.ParseFileDescriptorP)

	go func() {
		listener, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			logger.Fatal("new grpc listener err", zap.Error(err))
		}

		logger.Info("grpc server trying to listen on " + grpcAddr)
		if err = server.Serve(listener); err != nil {
			logger.Fatal("grpc server err", zap.Error(err))
		}
	}()

	return server
}

var _ pb.HelloServiceServer = (*helloServer)(nil)

type helloServer struct {
	pb.UnimplementedHelloServiceServer
}

func (h *helloServer) Unary(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	journalID, _ := vv.JournalID(ctx)
	logger.Info("unary receive: "+req.String(), zap.String("journal_id", journalID))

	if req.Message == "error" {
		err := errors.New("a dummy error occurs")
		return nil, vv.Error(codes.Internal, err.Error(), err)
	}

	if req.Message == "panic" {
		panic("a dummy panic occurs")
	}

	return &pb.HelloReply{
		SerialKey: "0123456789",
		Message:   time.Now().Format(time.RFC3339Nano),
	}, nil
}

func (h *helloServer) Stream(stream pb.HelloService_StreamServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return vv.Error(codes.Internal, err.Error(), errors.WithStack(err))
		}

		logger.Info("stream recv: " + req.String())

		if err := stream.Send(&pb.HelloReply{
			Message: req.Message + " Ack",
		}); err != nil {
			return vv.Error(codes.Internal, err.Error(), errors.WithStack(err))
		}
	}

	return nil
}

var _ pb.DummyServiceServer = (*dummyServer)(nil)

type dummyServer struct {
	pb.UnimplementedDummyServiceServer
}

func (d *dummyServer) Signup(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	fmt.Printf(">>>>>>>>> Signup Userinfo: %+v\n", vv.Userinfo(ctx))

	if req.Message == "error" {
		err := errors.New("a dummy error occurs")
		return nil, vv.Error(codes.Internal, err.Error(), err)
	}

	if req.Message == "panic" {
		panic("a dummy panic occurs")
	}

	if req.Message == "timeout" {
		time.Sleep(time.Second * 3)
	}

	return &pb.HelloReply{
		SerialKey: "0123456789",
		Message:   "ACK :" + req.Message + " @" + time.Now().Format(time.RFC3339Nano),
	}, nil
}

func (d *dummyServer) Dummy(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	fmt.Printf(">>>>>>>>> Dummy Userinfo: %+v\n", vv.Userinfo(ctx))

	time.Sleep(time.Millisecond * 50)

	return &pb.HelloReply{
		SerialKey: req.TrackId,
		Message:   "ACK :" + req.Message + " @" + time.Now().Format(time.RFC3339Nano),
	}, nil
}
