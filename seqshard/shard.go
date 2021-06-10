package seqshard

import (
	"context"
	pb "github.com/nathanieltornow/ostracon/seqshard/seqshardpb"
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

type committedRecord struct {
	gsn    int64
	record string
}

type Shard struct {
	pb.UnimplementedShardServer
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
	incomingComRec    chan *committedRecord
	comRecStreams     []chan *committedRecord
	comRecStreamsMu   sync.RWMutex
}

func NewShard(isRoot bool, batchingIntervall time.Duration) (*Shard, error) {
	newShard := &Shard{}
	newShard.isRoot = isRoot
	newShard.batchingIntervall = batchingIntervall
	newShard.streamToOR = make(map[pb.Shard_GetOrderServer]chan *orderResponse)
	newShard.snToPendingOR = make(map[int64]*orderRequest)
	newShard.incomingOR = make(chan *orderRequest, 4096)
	newShard.incomingComRec = make(chan *committedRecord, 4096)
	newShard.comRecStreams = make([]chan *committedRecord, 0)
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
	go s.broadcastComRec()
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
