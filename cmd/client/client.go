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
)

func main() {
	flag.Parse()
	conn, err := grpc.Dial(*parentIpAddr, grpc.WithInsecure())
	if err != nil {
		logrus.Fatalln("Failed making connection to seqshard")
	}
	defer conn.Close()

	shardClient := pb.NewRecordShardClient(conn)
	//time.Sleep(3 * time.Second)
	i := 1
	timeSum := time.Duration(0)
	for {
		if i > 9000 {
			break
		}
		time.Sleep(3 * time.Second)
		start := time.Now()
		res, err := shardClient.Append(context.Background(), &pb.AppendRequest{Record: "Hallo", Color: 1})
		appendTime := time.Since(start)
		timeSum += appendTime

		fmt.Println(res, appendTime)

		time.Sleep(3 * time.Second)

		start = time.Now()
		res, err = shardClient.Append(context.Background(), &pb.AppendRequest{Record: "Hallo", Color: 0})
		appendTime = time.Since(start)
		timeSum += appendTime

		fmt.Println(res, appendTime)
		if err != nil {
			logrus.Fatalln(err)
			return
		}
		i++
	}
	fmt.Println(timeSum)
	micros := timeSum.Microseconds()
	fmt.Println(micros / int64(i))

}
