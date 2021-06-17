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

type lsnColorTuple struct {
	lsn   int64
	color int64
}

type prevCurTuple struct {
	prev int64
	cur  int64
}

type record struct {
	gsn    chan int64
	record string
	color  int64
}

type RecordShard struct {
	rpb.UnimplementedRecordShardServer
	id               int64
	disk             *storage.ColoredStorage
	parentClient     *spb.ShardClient
	batchingInterval time.Duration

	colorToPrevLsn map[int64]int64
	curLsnMu       sync.Mutex

	colorToLsnTuple   map[int64]int64
	colorToLsnTupleMu sync.Mutex

	writeC chan *record

	lsnToRecord   map[lsnColorTuple]*record
	lsnToRecordMu sync.Mutex

	subscribeCs      []chan *rpb.CommittedRecord
	subscriberCsMu   sync.RWMutex
	newComRecC       chan *spb.CommittedRecord
	parentReadStream *spb.Shard_ReportCommittedRecordsClient
	replicaC         chan *rpb.CommittedRecord
}

func NewRecordShard(id int64, diskPath string, batchingInterval time.Duration) (*RecordShard, error) {
	disk, err := storage.NewColoredStorage(diskPath)
	if err != nil {
		return nil, err
	}
	s := new(RecordShard)
	s.id = id
	s.disk = disk
	s.batchingInterval = batchingInterval
	s.writeC = make(chan *record, 4096)
	s.lsnToRecord = make(map[lsnColorTuple]*record)
	s.subscribeCs = make([]chan *rpb.CommittedRecord, 0)
	s.newComRecC = make(chan *spb.CommittedRecord, 4096)
	s.replicaC = make(chan *rpb.CommittedRecord, 4096)
	s.colorToPrevLsn = make(map[int64]int64)
	s.colorToLsnTuple = make(map[int64]int64)
	s.colorToPrevLsn[0] = -1
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
		err = rs.connectToParent(parentIpAddr)
		if err != nil {
			return err
		}
	}
	logrus.Infoln("Starting RecordShard")

	heartBeatTick := time.Tick(20 * time.Second)
	go func() {
		for {
			<-heartBeatTick
			sum := int64(0)
			for i, _ := range rs.colorToPrevLsn {
				p, _ := rs.disk.GetCurrentLsn(i, true)
				sum += p + 1
			}
			logrus.Infof("Heartbeat: Stored %v records", sum)
		}
	}()

	if err := grpcServer.Serve(lis); err != nil {
		return err
	}

	return nil
}

func (rs *RecordShard) connectToParent(parentIpAddr string) error {
	conn, err := grpc.Dial(parentIpAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	client := spb.NewShardClient(conn)
	rs.parentClient = &client

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
