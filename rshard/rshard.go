package rshard

import (
	"context"
	"fmt"
	pb "github.com/nathanieltornow/ostracon/rshard/rshardpb"
	spb "github.com/nathanieltornow/ostracon/seqshard/seqshardpb"
	"google.golang.org/grpc"
	"net"
	"sync"
	"time"
)

type RecordShard struct {
	sync.Mutex
	pb.UnimplementedRecordShardServer

	orderStream spb.Shard_GetOrderClient
	subsStream  spb.Shard_ReportCommittedRecordsClient
	interval    time.Duration

	storagePath    string
	colorToService map[int64]*colorService
	orderReqC      chan *spb.OrderRequest
	comRecC        chan *spb.CommittedRecord
}

func NewRecordShard(storagePath string, interval time.Duration) (*RecordShard, error) {
	rs := new(RecordShard)
	rs.colorToService = make(map[int64]*colorService)
	rs.orderReqC = make(chan *spb.OrderRequest, 4096)
	rs.comRecC = make(chan *spb.CommittedRecord, 4096)
	rs.storagePath = storagePath
	rs.interval = interval
	return rs, nil
}

func (rs *RecordShard) Start(ipAddr, parentIpAddr string) error {
	lis, err := net.Listen("tcp", ipAddr)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	pb.RegisterRecordShardServer(grpcServer, rs)

	if parentIpAddr != "" {
		// wait for parent to be set up
		time.Sleep(3 * time.Second)
		err = rs.connectToParent(parentIpAddr)
		if err != nil {
			return err
		}
		fmt.Println("Connected to SeqShard on", parentIpAddr)
	}
	_, err = rs.addColor(0)
	if err != nil {
		return fmt.Errorf("couldn't start default color-service")
	}
	fmt.Println("Starting RecordShard")
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
	orderStream, err := client.GetOrder(context.Background())
	if err != nil {
		return err
	}
	rs.orderStream = orderStream

	subsStream, err := client.ReportCommittedRecords(context.Background())
	if err != nil {
		return err
	}
	rs.subsStream = subsStream

	var retErr error
	go func() {
		err := rs.sendOrderRequests()
		if err != nil {
			retErr = err
		}
	}()
	go func() {
		err := rs.receiveOrderResponses()
		if err != nil {
			retErr = err
		}
	}()
	go func() {
		err := rs.sendCommittedRecords()
		if err != nil {
			retErr = err
		}
	}()
	go func() {
		err := rs.receiveCommittedRecords()
		if err != nil {
			retErr = err
		}
	}()
	return retErr
}

func (rs *RecordShard) addColor(color int64) (*colorService, error) {
	cs, err := newColorService(fmt.Sprintf("%v/%v", rs.storagePath, color), color, rs.orderReqC, rs.comRecC, rs.interval)
	rs.Lock()
	_, ok := rs.colorToService[color]
	if ok {
		return nil, fmt.Errorf("color already in use")
	}
	rs.colorToService[color] = cs
	rs.Unlock()
	if err != nil {
		return cs, err
	}
	go func() {
		err = cs.Start()
	}()
	return cs, err
}
