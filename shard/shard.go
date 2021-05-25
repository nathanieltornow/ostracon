package shard

import (
	"context"
	pb "github.com/nathanieltornow/ostracon/shard/shardpb"
	"github.com/nathanieltornow/ostracon/storage"
	"github.com/sirupsen/logrus"
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
		time.Sleep(time.Second * 3)
		err = s.ConnectToNewParent(parentIpAddr)
		if err != nil {
			return err
		}
		logrus.Infoln("Connected to Parent", parentIpAddr)

	}

	logrus.Infoln("Starting shardserver")
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

	stream, err := client.GetOrder(context.Background())
	if err != nil {
		return err
	}
	go s.SendOrderRequests(stream)
	go s.ReceiveOrderResponses(stream)
	return nil
}

func (s *Shard) SendOrderRequests(stream pb.Shard_GetOrderClient) {
	lsnCopy := s.lsn
	for range time.Tick(s.batchingIntervall) {
		logrus.Infoln("ticking")
		if lsnCopy == s.lsn {
			continue
		}
		orderReq := pb.OrderRequest{StartLsn: lsnCopy, NumOfRecords: s.lsn - lsnCopy}
		logrus.Infoln("Sending OrderRequest")
		err := stream.Send(&orderReq)

		if err != nil {
			logrus.Fatalln("Failed to send order requests")
		}
	}
}

func (s *Shard) ReceiveOrderResponses(stream pb.Shard_GetOrderClient) {

	for {
		in, err := stream.Recv()
		logrus.Infoln("Received", in)
		if err != nil {
			logrus.Fatalln("Failed to receive order requests")
		}
		s.waitMap[in.StartLsn] <- in.StartGsn
	}

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
