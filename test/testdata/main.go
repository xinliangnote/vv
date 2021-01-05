package main

import (
	"context"
	"time"

	"github.com/koketama/auth"
	"github.com/koketama/shutdown"
	"github.com/koketama/zaplog"
	"github.com/spf13/cobra"
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
	root := &cobra.Command{
		Use: "vv",
	}

	var logfile string
	root.PersistentFlags().StringVar(&logfile, "logfile", "", "log file")

	root.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		var err error
		if logger, err = zaplog.NewJSONLogger(zaplog.WithInfoLevel(), zaplog.WithFileP(logfile)); err != nil {
			panic(err)
		}
	}

	root.AddCommand(
		normalCmd(),
		gatewayCmd(),
		serverCmd(),
		restCmd(),
	)

	if err := root.Execute(); err != nil {
		logger.Fatal("execute root cmd err", zap.Error(err))
	}
}

func normalCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "normal",
		Short: "the simple demo",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())

			server := newServer("127.0.0.1:7070", ":7080", "")
			gateway := newGateway(ctx, "127.0.0.1:7070", ":8080")

			time.Sleep(time.Second)
			newClient("127.0.0.1:7070")

			time.Sleep(time.Second)
			newRest("127.0.0.1:8080")

			time.Sleep(time.Second)
			restDummy(ctx, "127.0.0.1:8080", 1, 10)

			shutdown.NewHook().Close(
				func() {
					gateway.Shutdown(context.TODO())
					server.GracefulStop()
					cancel()
				},
			)
		},
	}
}

func gatewayCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gateway",
		Short: "grpc service gateway",
	}

	var (
		gatewayAddr string
		serverAddr  string
	)
	cmd.Flags().StringVar(&gatewayAddr, "gateway", "", "gateway addr")
	cmd.Flags().StringVar(&serverAddr, "server", "", "server addr")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		gateway := newGateway(context.TODO(), serverAddr, gatewayAddr)
		shutdown.NewHook().Close(
			func() {
				gateway.Shutdown(context.TODO())
			},
		)
	}

	return cmd
}

func serverCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "grpc service server",
	}

	var (
		serverAddr      string
		prometheusAddr  string
		pushgatewayAddr string
	)
	cmd.Flags().StringVar(&serverAddr, "server", "", "server addr")
	cmd.Flags().StringVar(&prometheusAddr, "prometheus", "", "prometheus addr")
	cmd.Flags().StringVar(&pushgatewayAddr, "pushgateway", "", "pushgateway addr")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		server := newServer(serverAddr, prometheusAddr, pushgatewayAddr)
		shutdown.NewHook().Close(
			func() {
				server.GracefulStop()
			},
		)
	}

	return cmd
}

func restCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rest",
		Short: "rest client",
	}

	var (
		gatewayAddr string
		goroutines  int
	)
	cmd.Flags().StringVar(&gatewayAddr, "gateway", "", "gateway addr")
	cmd.Flags().IntVar(&goroutines, "goroutines", 20, "goroutines")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		restDummy(context.TODO(), gatewayAddr, goroutines, 16<<10)
	}

	return cmd
}
