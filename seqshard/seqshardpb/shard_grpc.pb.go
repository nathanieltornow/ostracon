// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package seqshardpb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ShardClient is the client API for Shard service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ShardClient interface {
	GetOrder(ctx context.Context, opts ...grpc.CallOption) (Shard_GetOrderClient, error)
	ReportCommittedRecords(ctx context.Context, opts ...grpc.CallOption) (Shard_ReportCommittedRecordsClient, error)
}

type shardClient struct {
	cc grpc.ClientConnInterface
}

func NewShardClient(cc grpc.ClientConnInterface) ShardClient {
	return &shardClient{cc}
}

func (c *shardClient) GetOrder(ctx context.Context, opts ...grpc.CallOption) (Shard_GetOrderClient, error) {
	stream, err := c.cc.NewStream(ctx, &Shard_ServiceDesc.Streams[0], "/sshardpb.Shard/GetOrder", opts...)
	if err != nil {
		return nil, err
	}
	x := &shardGetOrderClient{stream}
	return x, nil
}

type Shard_GetOrderClient interface {
	Send(*OrderRequest) error
	Recv() (*OrderResponse, error)
	grpc.ClientStream
}

type shardGetOrderClient struct {
	grpc.ClientStream
}

func (x *shardGetOrderClient) Send(m *OrderRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *shardGetOrderClient) Recv() (*OrderResponse, error) {
	m := new(OrderResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *shardClient) ReportCommittedRecords(ctx context.Context, opts ...grpc.CallOption) (Shard_ReportCommittedRecordsClient, error) {
	stream, err := c.cc.NewStream(ctx, &Shard_ServiceDesc.Streams[1], "/sshardpb.Shard/ReportCommittedRecords", opts...)
	if err != nil {
		return nil, err
	}
	x := &shardReportCommittedRecordsClient{stream}
	return x, nil
}

type Shard_ReportCommittedRecordsClient interface {
	Send(*CommittedRecord) error
	Recv() (*CommittedRecord, error)
	grpc.ClientStream
}

type shardReportCommittedRecordsClient struct {
	grpc.ClientStream
}

func (x *shardReportCommittedRecordsClient) Send(m *CommittedRecord) error {
	return x.ClientStream.SendMsg(m)
}

func (x *shardReportCommittedRecordsClient) Recv() (*CommittedRecord, error) {
	m := new(CommittedRecord)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ShardServer is the server API for Shard service.
// All implementations must embed UnimplementedShardServer
// for forward compatibility
type ShardServer interface {
	GetOrder(Shard_GetOrderServer) error
	ReportCommittedRecords(Shard_ReportCommittedRecordsServer) error
	mustEmbedUnimplementedShardServer()
}

// UnimplementedShardServer must be embedded to have forward compatible implementations.
type UnimplementedShardServer struct {
}

func (UnimplementedShardServer) GetOrder(Shard_GetOrderServer) error {
	return status.Errorf(codes.Unimplemented, "method GetOrder not implemented")
}
func (UnimplementedShardServer) ReportCommittedRecords(Shard_ReportCommittedRecordsServer) error {
	return status.Errorf(codes.Unimplemented, "method ReportCommittedRecords not implemented")
}
func (UnimplementedShardServer) mustEmbedUnimplementedShardServer() {}

// UnsafeShardServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ShardServer will
// result in compilation errors.
type UnsafeShardServer interface {
	mustEmbedUnimplementedShardServer()
}

func RegisterShardServer(s grpc.ServiceRegistrar, srv ShardServer) {
	s.RegisterService(&Shard_ServiceDesc, srv)
}

func _Shard_GetOrder_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ShardServer).GetOrder(&shardGetOrderServer{stream})
}

type Shard_GetOrderServer interface {
	Send(*OrderResponse) error
	Recv() (*OrderRequest, error)
	grpc.ServerStream
}

type shardGetOrderServer struct {
	grpc.ServerStream
}

func (x *shardGetOrderServer) Send(m *OrderResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *shardGetOrderServer) Recv() (*OrderRequest, error) {
	m := new(OrderRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Shard_ReportCommittedRecords_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ShardServer).ReportCommittedRecords(&shardReportCommittedRecordsServer{stream})
}

type Shard_ReportCommittedRecordsServer interface {
	Send(*CommittedRecord) error
	Recv() (*CommittedRecord, error)
	grpc.ServerStream
}

type shardReportCommittedRecordsServer struct {
	grpc.ServerStream
}

func (x *shardReportCommittedRecordsServer) Send(m *CommittedRecord) error {
	return x.ServerStream.SendMsg(m)
}

func (x *shardReportCommittedRecordsServer) Recv() (*CommittedRecord, error) {
	m := new(CommittedRecord)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Shard_ServiceDesc is the grpc.ServiceDesc for Shard service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Shard_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "sshardpb.Shard",
	HandlerType: (*ShardServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetOrder",
			Handler:       _Shard_GetOrder_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "ReportCommittedRecords",
			Handler:       _Shard_ReportCommittedRecords_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "shard.proto",
}
