package main

import (
	"context"
	"flag"
	"fmt"
	pb "github.com/nathanieltornow/ostracon/rshard/rshardpb"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"time"
)

var (
	parentIpAddr = flag.String("parentIpAddr", "", "")
	gsn          = flag.Int64("gsn", 0, "as")
)

func main() {
	flag.Parse()
	conn, err := grpc.Dial(*parentIpAddr, grpc.WithInsecure())
	if err != nil {
		logrus.Fatalln("Failed making connection to seqshard")
	}
	defer conn.Close()

	shardClient := pb.NewRecordShardClient(conn)
	time.Sleep(3 * time.Second)

	stream, err := shardClient.Subscribe(context.Background(), &pb.ReadRequest{Gsn: *gsn, Color: 0})
	if err != nil {
		logrus.Fatalln(err)
	}
	for {
		in, err := stream.Recv()
		if err != nil {
			logrus.Fatalln(err)
		}
		fmt.Println(in)
	}
}
