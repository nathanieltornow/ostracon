package ostracon

import (
	"context"
	"flag"
	pb "github.com/nathanieltornow/ostracon/recordshard/recordshardpb"
	"google.golang.org/grpc"
	"os"
	"testing"
)

var (
	recShardIp = flag.String("ipaddr", "", "")
)

func TestMain(m *testing.M) {
	flag.Parse()
	ret := m.Run()
	os.Exit(ret)
}

func BenchmarkParallelRecordShard(b *testing.B) {
	conn, err := grpc.Dial(*recShardIp, grpc.WithInsecure())
	if err != nil {
		b.Errorf("Failed making connection to seqshard")
	}
	defer conn.Close()
	shardClient := pb.NewRecordShardClient(conn)
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			_, err = shardClient.Append(context.Background(), &pb.AppendRequest{Record: "Hallo"})
			if err != nil {
				b.Errorf("failed to append")
			}
		}
	})
}
