package main

import (
	"context"
	pb "github.com/nathanieltornow/ostracon/shard/shardpb"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"io"
	"time"
)

func main() {

	conn, err := grpc.Dial("localhost:4000", grpc.WithInsecure())
	if err != nil {
		logrus.Fatalln("Failed making connection to shard")
	}
	defer conn.Close()

	shardClient := pb.NewShardClient(conn)

	stream, err := shardClient.GetOrder(context.Background())

	waitc := make(chan struct{})

	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				logrus.Fatalf("%v", err)
			}
			if in.StartLsn != in.StartGsn {
				logrus.Fatalf("Got wrong StartLsn and StartGsn: %v, %v", in.StartLsn, in.StartGsn)
			}
		}
	}()

	i := int64(0)

	for range time.Tick(time.Second) {
		if err := stream.Send(&pb.OrderRequest{StartLsn: i, NumOfRecords: 1}); err != nil {
			logrus.Fatalf("Failed to send Order Request %v", err)
		}
		i++
	}
	stream.CloseSend()
	<-waitc
}
