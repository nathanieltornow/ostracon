package main

import (
	"context"
	pb "github.com/nathanieltornow/ostracon/shard/shardpb"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"time"
)

func main() {
	conn, err := grpc.Dial("localhost:4000", grpc.WithInsecure())
	if err != nil {
		logrus.Fatalln("Failed making connection to shard")
	}
	defer conn.Close()

	shardClient := pb.NewShardClient(conn)

	for range time.Tick(5 * time.Second) {
		record, err := shardClient.Append(context.Background(), &pb.AppendRequest{Record: "Hallo"})
		if err != nil {
			return
		}
		logrus.Println(record)
	}
}
