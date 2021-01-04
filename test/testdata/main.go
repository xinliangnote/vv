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
	if logger, err = zaplog.NewJSONLogger(zaplog.WithInfoLevel(), zaplog.WithFileP("/tmp/vv/log")); err != nil {
		panic(err)
	}
	defer logger.Sync()

	ctx, cancel := context.WithCancel(context.Background())

	server := newServer("127.0.0.1:7070", ":7080")
	gateway := newGateway(ctx, "127.0.0.1:7070", ":8080")

	time.Sleep(time.Second)
	newClient("127.0.0.1:7070")

	time.Sleep(time.Second)
	newRest("127.0.0.1:8080")

	time.Sleep(time.Second)
	restDummy(ctx, "127.0.0.1:8080")

	shutdown.NewHook().Close(
		func() {
			gateway.Shutdown(context.TODO())
			server.GracefulStop()
			cancel()
		},
	)
}
