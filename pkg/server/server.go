package server

import (
	"github.com/go-logr/logr"
	"google.golang.org/grpc"

	"sigs.k8s.io/controller-runtime/pkg/client"

	pipelinesv1 "github.com/bigkevmcd/peanut-helmpipelines/pkg/protos/pipelines/v1"
)

// NewGRPCServer creates a new gRPC server.
func NewGRPCServer(l logr.Logger, c client.Client, opts ...grpc.ServerOption) *grpc.Server {
	gsrv := grpc.NewServer(opts...)
	pipelinesv1.RegisterPipelinesServiceServer(gsrv, NewPipelinesServer(l, c))
	return gsrv
}
