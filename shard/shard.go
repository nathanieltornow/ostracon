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
	parentConn        *grpc.ClientConn
	parentClient      *pb.ShardClient
	waitMapMu         sync.Mutex
	waitMap           map[int64]chan int64 // maps sn to pending gsn
	isRoot            bool
	isSequencer       bool
	sn                int64
	snMu              sync.Mutex
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
		time.Sleep(2 * time.Second)
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
	lsnCopy := s.sn
	for range time.Tick(s.batchingIntervall) {
		if lsnCopy == s.sn {
			continue
		}
		orderReq := pb.OrderRequest{StartLsn: lsnCopy, NumOfRecords: s.sn - lsnCopy}
		err := stream.Send(&orderReq)

		lsnCopy = s.sn
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

		s.waitMapMu.Lock()
		for i := int64(0); i < in.NumOfRecords; i++ {
			if _, ok := s.waitMap[in.StartLsn+i]; ok {
				s.waitMap[in.StartLsn+i] <- in.StartGsn + i
			} else {
				logrus.Fatalln("Not in map")
			}
		}
		s.waitMapMu.Unlock()
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
