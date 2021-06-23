package main

import (
	"context"
	"flag"
	"fmt"
	pb "github.com/nathanieltornow/ostracon/recshard/recshardpb"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
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

var (
	read = flag.Bool("read", false, "")
)

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
	flag.Parse()
	if *read {
		err = subscribe(t.ShardIps[1], 100)
		if err != nil {
			panic(err)
		}
		return
	}

	interval := time.Duration(time.Second.Nanoseconds() / int64(t.Ops))

	for j := 15; j < 100; j++ {
		resC := make(chan *bResult, len(t.ShardIps))
		for i := 0; i < j; i++ {
			go appendBenchmark(t.ShardIps[0], false, t.Runtime, interval, resC)
		}
		ovrOps := 0
		ovrLat := time.Duration(0)
		for i := 0; i < j; i++ {
			res := <-resC
			ovrOps += res.operations
			ovrLat += res.overallLatency
		}
		ovrLat = time.Duration(ovrLat.Nanoseconds() / int64(j))
		fmt.Printf("Appended %v records in %v seconds with an average latency of %v\n", ovrOps, t.Runtime, ovrLat)
		fmt.Printf("Throughput: %v ops/sec\n", float64(ovrOps)/t.Runtime.Seconds())

		f, err := os.OpenFile("result.csv",
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
		}
		defer f.Close()
		if _, err := f.WriteString(fmt.Sprintf("%v, %v\n", float64(ovrOps)/t.Runtime.Seconds(), ovrLat.Microseconds())); err != nil {
			log.Println(err)
		}
	}

}

func appendBenchmark(ipAddr string, isWriter bool, runtime time.Duration, interval time.Duration, resultC chan *bResult) {
	ticker := time.Tick(interval)
	conn, err := grpc.Dial(ipAddr, grpc.WithInsecure())
	if err != nil {
		logrus.Errorf("Failed making connection to shard")
	}
	defer conn.Close()
	shardClient := pb.NewRecordShardClient(conn)

	var latencysum time.Duration
	i := 0

	nextMinute := time.Now().Truncate(time.Minute).Add(time.Minute)
	fmt.Printf("Append benchmark scheduled for %v\n", nextMinute)

	// wait for benchmark to start
	<-time.After(time.Until(nextMinute))

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
				return
			}
			i++
			latencysum += time.Since(start)
		}
	}
	lat := time.Duration(latencysum.Nanoseconds() / int64(i))
	resultC <- &bResult{operations: i, overallLatency: lat}
}

type gsnTime struct {
	time time.Time
	gsn  int64
}

func subscribe(ipAddr string, toGsn int64) error {
	conn, err := grpc.Dial(ipAddr, grpc.WithInsecure())
	if err != nil {
		logrus.Errorf("Failed making connection to shard")
	}
	defer conn.Close()
	sClient := pb.NewRecordShardClient(conn)
	stream, err := sClient.Subscribe(context.Background(), &pb.ReadRequest{Gsn: 0, Color: 0})
	if err != nil {
		return err
	}
	nextMinute := time.Now().Truncate(time.Minute).Add(time.Minute)
	fmt.Printf("Read/Write benchmark scheduled for %v\n", nextMinute)

	<-time.After(time.Until(nextMinute))
	fmt.Printf("jh")
	startTime := time.Now()
	for {
		in, err := stream.Recv()
		if err != nil {
			return err
		}
		fmt.Println(in)
		if in.Gsn > toGsn {
			break
		}
	}
	fmt.Println(time.Since(startTime))
	return nil
}
