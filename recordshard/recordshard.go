package recordshard

import (
	"context"
	rpb "github.com/nathanieltornow/ostracon/recordshard/recordshardpb"
	spb "github.com/nathanieltornow/ostracon/shard/shardpb"
	storage2 "github.com/nathanieltornow/ostracon/storage"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"sync"
	"time"
)

type committedRecord struct {
	gsn    int64
	record string
}

type RecordShard struct {
	rpb.UnimplementedRecordShardServer
	disk             *storage2.Storage
	parentConn       *grpc.ClientConn
	parentClient     *spb.ShardClient
	batchingInterval time.Duration
	waitMap          map[int64]chan int64
	waitMapMu        sync.Mutex
}

func (rs *RecordShard) Append(ctx context.Context, request *rpb.AppendRequest) (*rpb.CommittedRecord, error) {
	if rs.parentConn == nil && rs.parentClient == nil {
		lsn, _ := rs.disk.Write(request.Record)
		_ = rs.disk.Sync()
		_ = rs.disk.Commit(lsn, lsn)
		_ = rs.disk.Sync()
		return &rpb.CommittedRecord{Gsn: lsn, Record: request.Record}, nil
	}
	return nil, nil
}

func NewRecordShard(diskPath string, batchingInterval time.Duration) (*RecordShard, error) {
	disk, err := storage2.NewStorage(diskPath)
	if err != nil {
		return nil, err
	}
	s := RecordShard{}
	s.disk = disk
	s.batchingInterval = batchingInterval
	return &s, nil
}

func (rs *RecordShard) Start(ipAddr string, parentIpAddr string) error {
	lis, err := net.Listen("tcp", ipAddr)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	rpb.RegisterRecordShardServer(grpcServer, rs)

	if parentIpAddr != "" {
		time.Sleep(3 * time.Second)
		err = rs.ConnectToParent(parentIpAddr)
		if err != nil {
			return err
		}
	}
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
	go rs.SendOrderRequests(stream)
	go rs.ReceiveOrderResponses(stream)
	return nil
}

func (rs *RecordShard) SendOrderRequests(stream spb.Shard_GetOrderClient) {
	lsnCopy := rs.disk.GetNextLsn()
	for range time.Tick(rs.batchingInterval) {
		lsn := rs.disk.GetNextLsn()
		if lsnCopy == lsn {
			continue
		}
		orderReq := spb.OrderRequest{StartLsn: lsnCopy, NumOfRecords: lsn - lsnCopy}
		err := stream.Send(&orderReq)

		lsnCopy = lsn
		if err != nil {
			logrus.Fatalln("Failed to send order requests")
		}

	}
}

func (rs *RecordShard) ReceiveOrderResponses(stream spb.Shard_GetOrderClient) {
	for {
		_ = rs.disk.Sync()
		in, err := stream.Recv()
		if err != nil {
			return
		}
		for i := int64(0); i < in.NumOfRecords; i++ {
			_ = rs.disk.Commit(in.StartLsn+i, in.StartGsn+i)
		}
		_ = rs.disk.Sync()
		rs.waitMapMu.Lock()
		for i := int64(0); i < in.NumOfRecords; i++ {
			rs.waitMap[in.StartLsn+i] <- in.StartGsn + i
			logrus.Fatalln("Not in map")
		}
		rs.waitMapMu.Unlock()
	}
}
