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
	diskMu            sync.Mutex
	parentConn        *grpc.ClientConn
	parentClient      *pb.ShardClient
	waitMapMu         sync.Mutex
	waitMap           map[int64]chan int64 // maps lsn to pending gsn
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

	newShard := &Shard{}
	newShard.disk = disk
	newShard.isRoot = isRoot
	newShard.isSequencer = isSequencer
	newShard.batchingIntervall = batchingIntervall
	newShard.waitMap = make(map[int64]chan int64)
	return newShard, nil
}

func (s *Shard) Start(ipAddr string, parentIpAddr string) error {
	lis, err := net.Listen("tcp", ipAddr)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	pb.RegisterShardServer(grpcServer, s)

	if !s.isRoot {
		err = s.ConnectToNewParent(parentIpAddr)
		if err != nil {
			return err
		}
	}

	if err := grpcServer.Serve(lis); err != nil {
		return err
	}
	return nil
}

func (s *Shard) ConnectToNewParent(parentIpAddr string) error {
	conn, err := grpc.Dial(parentIpAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	client := pb.NewShardClient(conn)
	s.parentClient = &client
	s.parentConn = conn
	return nil
}

func (s *Shard) SendOrderRequests() error {
	return nil
}

func (s *Shard) ReceiveOrderResponses() error {
	return nil
}

func (s *Shard) WaitForGsn(lsn int64) int64 {
	s.waitMapMu.Lock()
	gsnC := make(chan int64)
	s.waitMap[lsn] = gsnC
	s.waitMapMu.Unlock()

	gsn := <-gsnC
	s.waitMapMu.Lock()
	delete(s.waitMap, lsn)
	s.waitMapMu.Unlock()
	return gsn
}
