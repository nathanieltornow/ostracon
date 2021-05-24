package shard

import (
	"context"
	pb "github.com/nathanieltornow/ostracon/shard/shardpb"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"io"
	"os"
	"testing"
	"time"
)

const (
	shardIpAddr = "localhost:3223"
)

var (
	shardClient pb.ShardClient
)

func TestMain(m *testing.M) {
	shard, err := NewShard("tmp", true, true, time.Second)
	if err != nil {
		logrus.Fatalln("Failed creating shard")
	}
	logrus.Infof("Starting shard on %v", shardIpAddr)

	go func() {
		err := shard.Start(shardIpAddr)
		if err != nil {
			logrus.Fatalln("Failed starting shard")
		}
	}()
	time.Sleep(time.Second)

	conn, err := grpc.Dial(shardIpAddr, grpc.WithInsecure())
	if err != nil {
		logrus.Fatalln("Failed making connection to shard")
	}
	defer conn.Close()

	shardClient = pb.NewShardClient(conn)
	ret := m.Run()
	os.Exit(ret)
}

func TestSimpleOrderClient(t *testing.T) {
	stream, err := shardClient.GetOrder(context.Background())
	if err != nil {
		t.Errorf("%v", err)
	}

	waitc := make(chan struct{})

	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				t.Errorf("%v", err)
			}
			if in.StartLsn != in.StartGsn {
				t.Errorf("Got wrong StartLsn and StartGsn: %v, %v", in.StartLsn, in.StartGsn)
			}
		}
	}()

	time.Sleep(time.Second)
	for i := int64(0); i < 10; i++ {

		if err := stream.Send(&pb.OrderRequest{StartLsn: i, NumOfRecords: 1}); err != nil {
			t.Errorf("Failed to send Order Request")
		}
	}
	stream.CloseSend()
	<-waitc
}
