package main

import (
	"context"
	pb "github.com/nathanieltornow/ostracon/recordshard/recordshardpb"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"log"
	"time"
)

func main() {
	conn, err := grpc.Dial("localhost:6000", grpc.WithInsecure())
	if err != nil {
		logrus.Fatalln("Failed making connection to shard")
	}
	defer conn.Close()

	shardClient := pb.NewRecordShardClient(conn)

	time.Sleep(3 * time.Second)
	for range time.Tick(3 * time.Second) {
		//fmt.Println("appending")
		start := time.Now()
		record, err := shardClient.Append(context.Background(), &pb.AppendRequest{Record: "Hallo"})
		appendTime := time.Since(start)
		if err != nil {
			logrus.Fatalln(err)
			return
		}
		log.Printf("Received %v with GSN %v; Time: %v", record.Record, record.Gsn, appendTime)
	}
}
