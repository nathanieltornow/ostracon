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

type orderResponse struct {
	startLsn     int64
	numOfRecords int64
	startGsn     int64
}

type orderRequest struct {
	startLsn     int64
	numOfRecords int64
	stream       pb.Shard_GetOrderServer
}

type Shard struct {
	pb.UnimplementedShardServer
	disk              *storage.Storage
	parentConn        *grpc.ClientConn
	parentClient      *pb.ShardClient
	isRoot            bool
	sn                int64
	snMu              sync.Mutex
	streamToOR        map[pb.Shard_GetOrderServer]chan *orderResponse
	streamToORMu      sync.Mutex
	snToPendingOR     map[int64]*orderRequest
	batchingIntervall time.Duration
	incomingOR        chan *orderRequest
}

func NewShard(diskPath string, isRoot bool, batchingIntervall time.Duration) (*Shard, error) {
	disk, err := storage.NewStorage(diskPath)
	if err != nil {
		return nil, err
	}

	newShard := &Shard{}
	newShard.disk = disk
	newShard.isRoot = isRoot
	newShard.batchingIntervall = batchingIntervall
	newShard.streamToOR = make(map[pb.Shard_GetOrderServer]chan *orderResponse)
	newShard.snToPendingOR = make(map[int64]*orderRequest)
	newShard.incomingOR = make(chan *orderRequest, 4096)
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
