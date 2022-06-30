// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// PipelinesServiceClient is the client API for PipelinesService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PipelinesServiceClient interface {
	// List all Pipelines
	ListPipelines(ctx context.Context, in *ListPipelinesRequest, opts ...grpc.CallOption) (*ListPipelinesResponse, error)
}

type pipelinesServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewPipelinesServiceClient(cc grpc.ClientConnInterface) PipelinesServiceClient {
	return &pipelinesServiceClient{cc}
}

func (c *pipelinesServiceClient) ListPipelines(ctx context.Context, in *ListPipelinesRequest, opts ...grpc.CallOption) (*ListPipelinesResponse, error) {
	out := new(ListPipelinesResponse)
	err := c.cc.Invoke(ctx, "/pipelines.v1.PipelinesService/ListPipelines", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PipelinesServiceServer is the server API for PipelinesService service.
// All implementations should embed UnimplementedPipelinesServiceServer
// for forward compatibility
type PipelinesServiceServer interface {
	// List all Pipelines
	ListPipelines(context.Context, *ListPipelinesRequest) (*ListPipelinesResponse, error)
}

// UnimplementedPipelinesServiceServer should be embedded to have forward compatible implementations.
type UnimplementedPipelinesServiceServer struct {
}

func (UnimplementedPipelinesServiceServer) ListPipelines(context.Context, *ListPipelinesRequest) (*ListPipelinesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListPipelines not implemented")
}

// UnsafePipelinesServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PipelinesServiceServer will
// result in compilation errors.
type UnsafePipelinesServiceServer interface {
	mustEmbedUnimplementedPipelinesServiceServer()
}

func RegisterPipelinesServiceServer(s *grpc.Server, srv PipelinesServiceServer) {
	s.RegisterService(&_PipelinesService_serviceDesc, srv)
}

func _PipelinesService_ListPipelines_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListPipelinesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PipelinesServiceServer).ListPipelines(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pipelines.v1.PipelinesService/ListPipelines",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PipelinesServiceServer).ListPipelines(ctx, req.(*ListPipelinesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _PipelinesService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "pipelines.v1.PipelinesService",
	HandlerType: (*PipelinesServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListPipelines",
			Handler:    _PipelinesService_ListPipelines_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pipelines/v1/pipelines_service.proto",
}
