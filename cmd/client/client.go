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
	i := 1
	timeSum := time.Duration(0)
	for range time.Tick(time.Millisecond * 500) {
		if i > 50 {
			break
		}
		//fmt.Println("appending")
		start := time.Now()
		_, err := shardClient.Append(context.Background(), &pb.AppendRequest{Record: "Hallo"})
		appendTime := time.Since(start)
		timeSum += appendTime
		if err != nil {
			logrus.Fatalln(err)
			return
		}
		i++
	}
	micros := timeSum.Microseconds()
	fmt.Println(micros / int64(i))

}
