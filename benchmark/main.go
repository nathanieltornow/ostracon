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
	read = flag.Int("read", 0, "")
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
	if *read > 0 {
		resultC := make(chan time.Duration, 1)
		readWrite(t.ShardIps[0], t.ShardIps[1], *read, 10, resultC)
		resDur := <-resultC
		f, err := os.OpenFile("read_result.csv",
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
		}
		defer f.Close()
		if _, err := f.WriteString(fmt.Sprintf("%v, %v\n", *read, resDur.Microseconds())); err != nil {
			log.Println(err)
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

type fromTo struct {
	from int64
	to   int64
}

func readWrite(ipAddr, readIpAddr string, numOfRecords, times int, resultC chan time.Duration) {
	conn, err := grpc.Dial(ipAddr, grpc.WithInsecure())
	if err != nil {
		logrus.Errorf("Failed making connection to shard")
	}
	defer conn.Close()
	shardClient := pb.NewRecordShardClient(conn)

	readerC := make(chan *fromTo, 64)
	waitC := make(chan bool)

	go subscribe(readIpAddr, readerC, waitC, resultC)
	// wait for benchmark to start
	//<-time.After(time.Until(nextMinute))

	time.Sleep(4 * time.Second)

	curGsn := int64(0)
	for j := 0; j < times; j++ {
		fmt.Println(j)
		from := curGsn
		for i := 0; i < numOfRecords; i++ {
			res, err := shardClient.Append(context.Background(), &pb.AppendRequest{Record: "Hallo", Color: 0})
			if err != nil {
				logrus.Errorf("failed to append")
				return
			}
			curGsn = res.Gsn
		}
		readerC <- &fromTo{from: from, to: curGsn}
	}
	close(readerC)
	<-waitC
	fmt.Println("h2")
}

func subscribe(ipAddr string, c chan *fromTo, finishedC chan bool, resultC chan time.Duration) error {
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
	times := time.Duration(0)

	k := 0
	curGsn := int64(0)
	for fT := range c {
		fmt.Println(fT.from, curGsn)
		if fT.from < curGsn {
			log.Fatalln("Oh no")
		}
		start := time.Now()
		for i := fT.from; i < fT.to; {
			in, err := stream.Recv()
			if err != nil {
				fmt.Println(err)
				return err
			}
			curGsn = in.Gsn
			i = curGsn
		}
		times += time.Since(start)
		k++
	}
	fmt.Println("hi")
	lat := time.Duration(times.Nanoseconds() / int64(k))
	resultC <- lat
	finishedC <- true
	return nil
}
