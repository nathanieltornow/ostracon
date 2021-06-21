package client

import (
	"context"
	"fmt"
	pb "github.com/nathanieltornow/ostracon/recshard/recshardpb"
	"github.com/nathanieltornow/ostracon/util"
	"google.golang.org/grpc"
	"math/rand"
	"time"
)

type Record struct {
	Record string
	Gsn    int64
	Color  int64
}

func Append(color int64, record string) (*Record, error) {
	ips, err := util.GetWriteShards(color)
	if err != nil {
		return nil, err
	}
	rand.Seed(time.Now().Unix())
	ip := ips[rand.Intn(len(ips))]

	conn, err := grpc.Dial(ip, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed making connection to recordshard")
	}
	defer conn.Close()

	client := pb.NewRecordShardClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.Append(ctx, &pb.AppendRequest{Record: record, Color: color})
	if err != nil || res == nil {
		return nil, fmt.Errorf("failed to append record\n")
	}
	return &Record{Record: res.Record, Gsn: res.Gsn, Color: res.Color}, nil

}

func Subscribe(gsn, color int64, resultC chan *Record) error {
	ipGroups, err := util.GetReadShards(color)
	if err != nil {
		return err
	}
	ips := make([]string, 0)
	for _, group := range ipGroups {
		rand.Seed(time.Now().Unix())
		ips = append(ips, group[rand.Intn(len(group))])
	}
	var retErr error
	for _, ip := range ips {
		ip2 := ip
		go func() {
			err := subscribeToSingleShard(ip2, gsn, color, resultC)
			if err != nil {
				retErr = err
			}
		}()
	}
	return retErr

}

func subscribeToSingleShard(ipAddr string, gsn, color int64, resultC chan *Record) error {
	conn, err := grpc.Dial(ipAddr, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to make connection: %v\n", err)
	}
	defer conn.Close()

	shardClient := pb.NewRecordShardClient(conn)
	stream, err := shardClient.Subscribe(context.Background(), &pb.ReadRequest{Gsn: gsn, Color: color})
	if err != nil {
		return err
	}
	for {
		in, err := stream.Recv()
		if err != nil {
			return err
		}
		resultC <- &Record{Record: in.Record, Gsn: in.Gsn, Color: color}
	}
}
