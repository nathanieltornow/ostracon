package seqshard

import (
	"context"
	pb "github.com/nathanieltornow/ostracon/seqshard/seqshardpb"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"io"
	"testing"
	"time"
)

func TestRootShard(t *testing.T) {
	shardIpAddr := "localhost:7000"
	shard, err := NewSeqShard(0, true, time.Second)
	if err != nil {
		t.Errorf("Failed creating seqshard")
	}

	go func() {
		err := shard.Start(shardIpAddr, "")
		if err != nil {
			t.Errorf("Failed starting seqshard")
		}
	}()
	time.Sleep(time.Second)

	conn, err := grpc.Dial(shardIpAddr, grpc.WithInsecure())
	if err != nil {
		t.Errorf("Failed making connection to seqshard")
	}
	defer conn.Close()

	shardClient := pb.NewShardClient(conn)
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
	for i := int64(0); i < 15; i++ {

		if err := stream.Send(&pb.OrderRequest{StartLsn: i, NumOfRecords: 1}); err != nil {
			t.Errorf("Failed to send Order Request")
		}
	}
	stream.CloseSend()
	<-waitc
}

func TestOrderChain(t *testing.T) {
	rootShardIpAddr := "localhost:3223"
	middleShardIpAddr := "localhost:3224"
	rootShard, err := NewSeqShard(0, true, time.Second)
	if err != nil {
		t.Errorf("Failed creating rootShard: %v", err)
	}

	go func() {
		err := rootShard.Start(rootShardIpAddr, "")
		if err != nil {
			t.Errorf("Failed starting rootShard")
		}
	}()
	time.Sleep(time.Second)

	middleShard, err := NewSeqShard(-1, false, time.Second)
	if err != nil {
		t.Errorf("Failed creating middleShard: %v", err)
	}
	go func() {
		err := middleShard.Start(middleShardIpAddr, rootShardIpAddr)
		if err != nil {
			t.Errorf("Failed starting middleShard: %v", err)
		}
	}()
	time.Sleep(time.Second)

	conn, err := grpc.Dial(middleShardIpAddr, grpc.WithInsecure())
	if err != nil {
		t.Errorf("Failed making connection to seqshard")
	}
	defer conn.Close()

	shardClient := pb.NewShardClient(conn)
	stream, err := shardClient.GetOrder(context.Background())
	if err != nil {
		t.Errorf("%v", err)
	}
	time.Sleep(time.Second)
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
			require.Equal(t, in.StartLsn, in.StartGsn)
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
