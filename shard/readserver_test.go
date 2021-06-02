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
	shardIpAddr := "localhost:6666"
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

	shardClient2 := pb.NewShardClient(conn)
	stream2, err := shardClient2.ReportCommittedRecords(context.Background())

	if err != nil {
		t.Errorf("%v", err)
	}
	go func() {
		for {
			in, err := stream2.Recv()
			fmt.Println(in)

			if in.Record == "hi" && in.Gsn == 14 {
				close(waitc)
				return
			} else {
				t.Errorf("Got wrong message")
			}
			if err == io.EOF {
				return
			}
			if err != nil {
				t.Errorf("%v", err)
			}

		}
	}()
	time.Sleep(time.Second)
	if err := stream1.Send(&pb.CommittedRecord{Gsn: 14, Record: "hi"}); err != nil {
		t.Errorf("Failed to send Order Request: %v", err)
	}
	stream1.CloseSend()

	select {
	case <-waitc:
	case <-time.After(time.Second * 2):
		t.Errorf("Timed out")

	}
}
