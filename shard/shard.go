package shard

import (
	pb "github.com/nathanieltornow/ostracon/shard/shardpb"
	"github.com/nathanieltornow/ostracon/storage"
	"google.golang.org/grpc"
	"net"
	"sync"
	"time"
)

type Shard struct {
	pb.UnimplementedShardServer
	disk              *storage.Storage
	parentClient      *pb.ShardClient
	isRoot            bool
	isSequencer       bool
	lsn               int64
	lsnMu             sync.Mutex
	batchingIntervall time.Duration
}

func NewShard(diskPath string, isRoot, isSequencer bool, batchingIntervall time.Duration) (*Shard, error) {
	disk, err := storage.NewStorage(diskPath)
	if err != nil {
		return nil, err
	}
	return &Shard{disk: disk, isRoot: isRoot, isSequencer: isSequencer, batchingIntervall: batchingIntervall}, nil
}

func (s *Shard) Start(ipAddr string) error {
	lis, err := net.Listen("tcp", ipAddr)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	pb.RegisterShardServer(grpcServer, s)

	if err := grpcServer.Serve(lis); err != nil {
		return err
	}
	return nil
}

func (s *Shard) SendOrderRequests() error {
	return nil
}

func (s *Shard) ReceiveOrderResponses() error {
	return nil
}
