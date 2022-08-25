package main

import (
	"log"
	"net"
	"os"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/go-logr/zapr"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/bigkevmcd/peanut-helmpipelines/pkg/server"
)

const (
	listenFlag = "listen"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(helmv2.AddToScheme(scheme))
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	viper.AutomaticEnv()
}

func makeRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "peanut-pipelines",
		Short: "Provides a gRPC API for parsing HelmReleases into pipelines",
		Run: func(cmd *cobra.Command, args []string) {
			zapLog, err := zap.NewDevelopment()
			cobra.CheckErr(err)
			logger := zapr.NewLogger(zapLog)

			cfg, err := config.GetConfig()
			cobra.CheckErr(err)

			cl, err := client.New(cfg, client.Options{Scheme: scheme})
			cobra.CheckErr(err)

			srv := server.NewGRPCServer(logger, cl,
				grpc.StreamInterceptor(
					grpc_middleware.ChainStreamServer(grpc_prometheus.StreamServerInterceptor),
				),
				grpc.UnaryInterceptor(
					grpc_middleware.ChainUnaryServer(
						grpc_prometheus.UnaryServerInterceptor,
						grpc_zap.UnaryServerInterceptor(zapLog),
					),
				),
			)
			reflection.Register(srv)

			log.Printf("Listening at %s", viper.GetString(listenFlag))
			lis, err := net.Listen("tcp", ":"+viper.GetString(listenFlag))
			cobra.CheckErr(err)
			cobra.CheckErr(srv.Serve(lis))
		},
	}

	cmd.Flags().String(
		listenFlag,
		portFromEnv(),
		"gRPC server listen port",
	)
	cobra.CheckErr(viper.BindPFlag(listenFlag, cmd.Flags().Lookup(listenFlag)))
	return cmd
}

func main() {
	cobra.CheckErr(makeRootCmd().Execute())
}

func portFromEnv() string {
	if v := os.Getenv("PORT"); v != "" {
		return v
	}
	return "8080"
}
