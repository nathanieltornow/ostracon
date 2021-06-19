package main

import (
	"context"
	"fmt"
	pb "github.com/nathanieltornow/ostracon/rshard/rshardpb"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"time"
)

type TestConfig struct {
	Ops      int           `yaml:"ops"`
	Runtime  time.Duration `yaml:"runtime"`
	ShardIps []string      `yaml:"shard_ips"`
}

type bResult struct {
	operations     int
	overallLatency time.Duration
}

func main() {
	buf, err := ioutil.ReadFile("benchmark/benchmark.config.yaml")
	if err != nil {
		logrus.Fatalln("failed to read file")
	}
	t := TestConfig{}
	err = yaml.Unmarshal(buf, &t)
	if err != nil {
		logrus.Fatalln("failed to unmarshal config")
	}

	interval := time.Duration(time.Second.Nanoseconds() / int64(t.Ops))

	resC := make(chan *bResult, len(t.ShardIps))
	for i := 0; i < len(t.ShardIps); i++ {
		go appendBenchmark(t.ShardIps[i], t.Runtime, interval, resC)
	}
	ovrOps := 0
	ovrLat := time.Duration(0)
	for i := 0; i < len(t.ShardIps); i++ {
		res := <-resC
		ovrOps += res.operations
		ovrLat += res.overallLatency
	}
	ovrLat = time.Duration(ovrLat.Nanoseconds() / int64(len(t.ShardIps)))
	fmt.Printf("Appended %v records in %v seconds with an average latency of %v\n", ovrOps, t.Runtime, ovrLat)
}

func appendBenchmark(ipAddr string, runtime time.Duration, interval time.Duration, resultC chan *bResult) {
	ticker := time.Tick(interval)
	conn, err := grpc.Dial(ipAddr, grpc.WithInsecure())
	if err != nil {
		logrus.Errorf("Failed making connection to shard")
	}
	defer conn.Close()
	shardClient := pb.NewRecordShardClient(conn)

	var latencysum time.Duration
	i := 0

	stop := time.After(runtime)
out:
	for {
		select {
		case <-stop:
			break out
		default:
			<-ticker
			start := time.Now()
			_, err = shardClient.Append(context.Background(), &pb.AppendRequest{Record: "Hallo", Color: 0})
			if err != nil {
				logrus.Errorf("failed to append")
			}
			i++
			latencysum += time.Since(start)
		}
	}
	lat := time.Duration(latencysum.Nanoseconds() / int64(i))
	resultC <- &bResult{operations: i, overallLatency: lat}
}
