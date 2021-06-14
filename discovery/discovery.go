package discovery

import (
	"context"
	pb "github.com/nathanieltornow/ostracon/discovery/discoverypb"
	"google.golang.org/grpc"
	"net"
	"sync"
)

type discoveryServer struct {
	sync.Mutex
	pb.DiscoveryServer
	shards []*pb.Shard
}

func NewDiscoveryServer() *discoveryServer {
	ds := new(discoveryServer)
	ds.shards = make([]*pb.Shard, 0)
	return ds
}

func (d *discoveryServer) Start(ipAddr string) error {
	lis, err := net.Listen("tcp", ipAddr)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	pb.RegisterDiscoveryServer(grpcServer, d)
	if err := grpcServer.Serve(lis); err != nil {
		return err
	}
	return nil
}

func (d *discoveryServer) Discover(_ context.Context, _ *pb.Empty) (*pb.DiscoveryResponse, error) {
	d.Lock()
	res := new(pb.DiscoveryResponse)
	res.Shards = d.shards
	d.Unlock()
	return res, nil
}

func (d *discoveryServer) RegisterShard(_ context.Context, s *pb.Shard) (*pb.Empty, error) {
	d.Lock()
	d.shards = append(d.shards, s)
	d.Unlock()
	return new(pb.Empty), nil
}
