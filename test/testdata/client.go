package main

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/bluekaki/vv/builder/client"
	"github.com/bluekaki/vv/test/testdata/pb/gen"

	"github.com/koketama/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func newClient() {
	conn, err := client.New(grpcAddr, client.WithSign(func(fullMethod string, message []byte) (authorization, date string, err error) {
		return signature.Generate("webapi", auth.MethodGRPC, fullMethod, message)
	}))
	if err != nil {
		logger.Fatal("new client err", zap.Error(err))
	}
	defer conn.Close()

	client := pb.NewHelloServiceClient(conn)

	callUnaryNormal(client)
	callUnaryError(client)
	callUnaryPanic(client)

	callStream(client)
}

func callUnaryNormal(client pb.HelloServiceClient) {
	fmt.Println("---------------------------------------------------------")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "minami")
	reply, err := client.Unary(ctx, &pb.HelloRequest{Message: "normal"},
		grpc.WaitForReady(true),
		grpc.UseCompressor(gzip.Name),
	)
	if err != nil {
		if s, ok := status.FromError(err); ok {
			logger.Error("call unary normal err", zap.Any("error", s.Proto().String()))

		} else {
			logger.Error("call unary normal err", zap.Error(err))
		}
		return
	}

	logger.Info("unary normal reply: " + reply.String())
}

func callUnaryError(client pb.HelloServiceClient) {
	fmt.Println("---------------------------------------------------------")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "minami")
	reply, err := client.Unary(ctx, &pb.HelloRequest{Message: "error"},
		grpc.WaitForReady(true),
		grpc.UseCompressor(gzip.Name),
	)
	if err != nil {
		if s, ok := status.FromError(err); ok {
			logger.Error("call unary error err", zap.Any("error", s.Proto().String()))

		} else {
			logger.Error("call unary error err", zap.Error(err))
		}
		return
	}

	logger.Info("unary error reply: " + reply.String())
}

func callUnaryPanic(client pb.HelloServiceClient) {
	fmt.Println("---------------------------------------------------------")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "minami")
	reply, err := client.Unary(ctx, &pb.HelloRequest{Message: "panic"},
		grpc.WaitForReady(true),
		grpc.UseCompressor(gzip.Name),
	)
	if err != nil {
		if s, ok := status.FromError(err); ok {
			logger.Error("call unary panic err", zap.Any("error", s.Proto().String()))

		} else {
			logger.Error("call unary panic err", zap.Error(err))
		}
		return
	}

	logger.Info("unary panic reply: " + reply.String())
}

func callStream(client pb.HelloServiceClient) {
	fmt.Println("---------------------------------------------------------")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "minami")
	stream, err := client.Stream(ctx,
		grpc.WaitForReady(true),
		grpc.UseCompressor(gzip.Name),
	)
	if err != nil {
		if s, ok := status.FromError(err); ok {
			logger.Error("call stream err", zap.Any("error", s.Proto().String()))

		} else {
			logger.Error("call stream err", zap.Error(err))
		}
		return
	}

	go func() {
		for _, v := range []string{"1", "2", "3", "4"} {
			if err := stream.Send(&pb.HelloRequest{
				Message: "hello stream " + v,
			}); err != nil {
				if err == io.EOF {
					return
				}

				if s, ok := status.FromError(err); ok {
					logger.Error("stream send err", zap.Any("error", s.Proto().String()))

				} else {
					logger.Error("stream send err", zap.Error(err))
				}
				return
			}
		}
		stream.CloseSend()
	}()

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			if s, ok := status.FromError(err); ok {
				logger.Error("stream recv err", zap.Any("error", s.Proto().String()))

			} else {
				logger.Error("stream recv err", zap.Error(err))
			}
			return
		}

		logger.Info("stream recv: " + resp.String())
	}
}
