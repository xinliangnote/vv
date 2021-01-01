package main

import (
	"context"
	"time"

	"github.com/koketama/auth"
	"github.com/koketama/shutdown"
	"github.com/koketama/zaplog"
	"go.uber.org/zap"
)

var logger *zap.Logger
var signature auth.Signature

const (
	grpcAddr = "127.0.0.1:7070"
	restAddr = ":8080"
)

func init() {
	var err error
	signature, err = auth.NewSignature(auth.WithSHA256(), auth.WithSecrets(map[auth.Identifier]auth.Secret{
		"webapi": "QZ74a6yb9tejrquz4yos",
	}))
	if err != nil {
		panic(err)
	}
}

func main() {
	var err error
	if logger, err = zaplog.NewJSONLogger(zaplog.WithInfoLevel(), zaplog.WithFileP("/tmp/grpcgw/log")); err != nil {
		panic(err)
	}
	defer logger.Sync()

	ctx, cancel := context.WithCancel(context.Background())

	server := newServer()
	gateway := newGateway(ctx)

	time.Sleep(time.Second)
	newClient()

	time.Sleep(time.Second)
	newRest()

	time.Sleep(time.Second)
	restDummy(ctx)

	shutdown.NewHook().Close(
		func() {
			gateway.Shutdown(context.TODO())
			server.GracefulStop()
			cancel()
		},
	)
}
