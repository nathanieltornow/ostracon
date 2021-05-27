package main

import (
	"context"
	"flag"
	"fmt"
	pb "github.com/nathanieltornow/ostracon/recordshard/recordshardpb"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"time"
)

var (
	parentIpAddr = flag.String("parentIpAddr", "", "")
)

func main() {
	flag.Parse()
	conn, err := grpc.Dial(*parentIpAddr, grpc.WithInsecure())
	if err != nil {
		logrus.Fatalln("Failed making connection to shard")
	}
	defer conn.Close()

	shardClient := pb.NewRecordShardClient(conn)
	time.Sleep(3 * time.Second)
	for range time.Tick(time.Second) {
		//fmt.Println("appending")
		start := time.Now()
		record, err := shardClient.Append(context.Background(), &pb.AppendRequest{Record: "Hallo"})
		appendTime := time.Since(start)
		if err != nil {
			logrus.Fatalln(err)
			return
		}
		fmt.Printf("Received %v with GSN %v; Time: %v", record.Record, record.Gsn, appendTime)
	}
}
