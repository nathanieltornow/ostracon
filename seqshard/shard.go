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
	color        int64
}

type orderRequest struct {
	startLsn     int64
	numOfRecords int64
	color        int64
	stream       pb.Shard_GetOrderServer
}

type committedRecord struct {
	gsn    int64
	record string
	color  int64
}

type snColorTuple struct {
	sn    int64
	color int64
}

type SeqShard struct {
	pb.UnimplementedShardServer
	color             int64
	isRoot            bool
	batchingIntervall time.Duration
	parentClient      *pb.ShardClient

	colorToSn   map[int64]int64
	colorToSnMu sync.Mutex

	sn               int64
	waitingOrderReqs map[snColorTuple]*orderRequest
	snMu             sync.Mutex

	orderRespCs   map[pb.Shard_GetOrderServer]chan *orderResponse
	orderRespCsMu sync.Mutex

	orderReqsC         chan *orderRequest
	comRecCIncoming    chan *committedRecord
	comRecCsOutgoing   []chan *committedRecord
	comRecCsOutgoingMu sync.RWMutex
}

func NewSeqShard(color int64, isRoot bool, batchingIntervall time.Duration) (*SeqShard, error) {
	newShard := &SeqShard{}
	newShard.isRoot = isRoot
	newShard.batchingIntervall = batchingIntervall
	newShard.orderRespCs = make(map[pb.Shard_GetOrderServer]chan *orderResponse)
	newShard.waitingOrderReqs = make(map[snColorTuple]*orderRequest)
	newShard.orderReqsC = make(chan *orderRequest, 4096)
	newShard.comRecCIncoming = make(chan *committedRecord, 4096)
	newShard.comRecCsOutgoing = make([]chan *committedRecord, 0)
	newShard.color = color
	newShard.colorToSn = make(map[int64]int64)
	return newShard, nil
}

// Start will start the SeqShard on the given IP-address and try to connect to its parent
func (s *SeqShard) Start(ipAddr string, parentIpAddr string) error {
	lis, err := net.Listen("tcp", ipAddr)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	pb.RegisterShardServer(grpcServer, s)

	if !s.isRoot {
		time.Sleep(3 * time.Second)
		err = s.connectToParent(parentIpAddr)
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

func (s *SeqShard) connectToParent(parentIpAddr string) error {
	conn, err := grpc.Dial(parentIpAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	client := pb.NewShardClient(conn)
	s.parentClient = &client
	stream, err := client.GetOrder(context.Background())
	if err != nil {
		return err
	}
	go s.sendOrderRequests(stream)
	go s.receiveOrderResponses(stream)
	return nil
}
