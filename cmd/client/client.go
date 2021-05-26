package main

import (
	"context"
	pb "github.com/nathanieltornow/ostracon/shard/shardpb"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"time"
)

func main() {
	conn, err := grpc.Dial("localhost:4001", grpc.WithInsecure())
	if err != nil {
		logrus.Fatalln("Failed making connection to shard")
	}
	defer conn.Close()

	shardClient := pb.NewShardClient(conn)

	time.Sleep(3 * time.Second)
	for range time.Tick(1 * time.Second) {
		logrus.Println("Appending...")
		record, err := shardClient.Append(context.Background(), &pb.AppendRequest{Record: "Hallo"})
		if err != nil {
			logrus.Fatalln(err)
			return
		}
		logrus.Println(record)
	}
}
