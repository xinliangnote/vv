package main

import (
	"context"
	"net/http"

	"github.com/bluekaki/vv/builder/gateway"
	"github.com/bluekaki/vv/test/testdata/pb/gen"

	"go.uber.org/zap"
)

func newGateway(ctx context.Context, grpcAddr, restAddr string) *http.Server {
	mux, options := gateway.New()
	if err := pb.RegisterDummyServiceHandlerFromEndpoint(ctx, mux, grpcAddr, options); err != nil {
		logger.Fatal("register gateway err", zap.Error(err))
	}

	server := &http.Server{
		Addr:    restAddr,
		Handler: mux,
	}

	go func() {
		logger.Info("gateway trying to listen on " + restAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("gateway err", zap.Error(err))
		}
	}()

	return server
}
