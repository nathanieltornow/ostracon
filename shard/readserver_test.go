package shard

import (
	"context"
	"fmt"
	pb "github.com/nathanieltornow/ostracon/shard/shardpb"
	"google.golang.org/grpc"
	"io"
	"testing"
	"time"
)

func TestRecBroadCast(t *testing.T) {
	// setup readserver
	shardIpAddr := "localhost:3223"
	shard, err := NewShard(true, time.Second)
	if err != nil {
		t.Errorf("Failed creating shard")
	}

	go func() {
		err := shard.Start(shardIpAddr, "")
		if err != nil {
			t.Errorf("Failed starting shard")
		}
	}()
	time.Sleep(time.Second)

	conn, err := grpc.Dial(shardIpAddr, grpc.WithInsecure())
	if err != nil {
		t.Errorf("Failed making connection to shard")
	}
	defer conn.Close()

	waitc := make(chan struct{})

	shardClient1 := pb.NewShardClient(conn)
	stream1, err := shardClient1.ReportCommittedRecords(context.Background())

	if err != nil {
		t.Errorf("%v", err)
	}
	go func() {
		for {
			in, err := stream1.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				t.Errorf("%v", err)
			}
			fmt.Println("1:", in)
		}
	}()

	shardClient2 := pb.NewShardClient(conn)
	stream2, err := shardClient2.ReportCommittedRecords(context.Background())

	if err != nil {
		t.Errorf("%v", err)
	}
	go func() {
		for {
			in, err := stream2.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				t.Errorf("%v", err)
			}
			fmt.Println("2:", in)
		}
	}()

	for range time.Tick(time.Second) {
		if err := stream1.Send(&pb.CommittedRecord{Gsn: 14, Record: "hi"}); err != nil {
			t.Errorf("Failed to send Order Request: %v", err)
		}
	}
	stream1.CloseSend()
	<-waitc
}
