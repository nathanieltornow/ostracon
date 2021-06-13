package recordshard

import (
	"context"
	rpb "github.com/nathanieltornow/ostracon/recordshard/recordshardpb"
	"github.com/nathanieltornow/ostracon/recordshard/storage"
	spb "github.com/nathanieltornow/ostracon/seqshard/seqshardpb"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"sync"
	"time"
)

type record struct {
	gsn    chan int64
	record string
}

type RecordShard struct {
	rpb.UnimplementedRecordShardServer
	id               int64
	disk             *storage.Storage
	parentConn       *grpc.ClientConn
	parentClient     *spb.ShardClient
	batchingInterval time.Duration
	curLsn           int64
	curLsnMu         sync.Mutex
	writeC           chan *record
	lsnToRecord      map[int64]*record
	lsnToRecordMu    sync.Mutex
	subscribeCs      []chan *rpb.CommittedRecord
	subscriberCsMu   sync.RWMutex
	newComRecC       chan *spb.CommittedRecord
	parentReadStream *spb.Shard_ReportCommittedRecordsClient
	replicaC         chan *rpb.CommittedRecord
}

func NewRecordShard(id int64, diskPath string, batchingInterval time.Duration) (*RecordShard, error) {
	disk, err := storage.NewStorage(diskPath, 0, 2, 10000000)
	if err != nil {
		return nil, err
	}
	s := new(RecordShard)
	s.id = id
	s.disk = disk
	s.batchingInterval = batchingInterval
	s.writeC = make(chan *record, 4096)
	s.lsnToRecord = make(map[int64]*record)
	s.curLsn = -1
	s.subscribeCs = make([]chan *rpb.CommittedRecord, 0)
	s.newComRecC = make(chan *spb.CommittedRecord, 4096)
	s.replicaC = make(chan *rpb.CommittedRecord, 4096)
	return s, nil
}

func (rs *RecordShard) Start(ipAddr string, parentIpAddr string) error {
	lis, err := net.Listen("tcp", ipAddr)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	rpb.RegisterRecordShardServer(grpcServer, rs)

	if parentIpAddr != "" {
		// wait for parent to be set up
		time.Sleep(3 * time.Second)
		err = rs.ConnectToParent(parentIpAddr)
		if err != nil {
			return err
		}
	}
	logrus.Infoln("Starting RecordShard")
	if err := grpcServer.Serve(lis); err != nil {
		return err
	}

	return nil
}

func (rs *RecordShard) ConnectToParent(parentIpAddr string) error {
	conn, err := grpc.Dial(parentIpAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	client := spb.NewShardClient(conn)
	rs.parentClient = &client
	rs.parentConn = conn

	stream, err := client.GetOrder(context.Background())
	if err != nil {
		return err
	}
	readStream, err := client.ReportCommittedRecords(context.Background())
	if err != nil {
		return err
	}
	go rs.writeAppends()
	go rs.sendOrderRequests(stream)
	go rs.receiveOrderResponses(stream)
	rs.parentReadStream = &readStream
	go rs.sendCommittedRecords(readStream)
	go rs.receiveCommittedRecords(readStream)
	go rs.storeReplicas()
	return nil
}
