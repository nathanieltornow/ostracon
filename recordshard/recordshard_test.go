package recordshard

import (
	"context"
	pb "github.com/nathanieltornow/ostracon/recordshard/recordshardpb"
	"github.com/nathanieltornow/ostracon/seqshard"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"math/rand"
	"os"
	"testing"
	"time"
)

const (
	shardIp1 = ":6000"
	shardIp2 = ":6001"
	letters  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func TestMain(m *testing.M) {
	shard, err := seqshard.NewSeqShard(0, true, time.Second)
	if err != nil {
		logrus.Fatalln("Failed creating seqshard")
	}
	go func() {
		err := shard.Start(":5000", "")
		if err != nil {
			logrus.Fatalln("Failed starting seqshard")
		}
	}()
	time.Sleep(time.Second)
	go func() {
		err := startRecordShard(shardIp1, ":5000")
		if err != nil {
			logrus.Fatalln("Failed starting recshard")
		}
	}()
	go func() {
		err := startRecordShard(shardIp2, ":5000")
		if err != nil {
			logrus.Fatalln("Failed starting recshard")
		}
	}()
	time.Sleep(3 * time.Second)
	ret := m.Run()
	os.Exit(ret)
}

func TestWrite(t *testing.T) {
	conn, err := grpc.Dial(shardIp1, grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()
	client := pb.NewRecordShardClient(conn)
	curGsn := int64(0)
	checkC := make(chan int64, 64)
	go checkDistinctGSN(t, checkC)
	for i := 0; i < 10; i++ {
		record := randString(10)
		res, err := client.Append(context.Background(), &pb.AppendRequest{
			Record: record,
			Color:  0,
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, res.Record, record)
		require.True(t, curGsn <= res.Gsn, "GSN has to increase")
		curGsn = res.Gsn
		checkC <- curGsn
	}
	time.Sleep(time.Second)
	close(checkC)
}

func TestConcurrentWrites(t *testing.T) {
	threads := 5
	waitC := make(chan bool, threads)
	checkC := make(chan int64, 64)
	go checkDistinctGSN(t, checkC)
	for i := 0; i < threads; i++ {
		go func() {
			conn, err := grpc.Dial(shardIp1, grpc.WithInsecure())
			require.NoError(t, err)
			defer conn.Close()
			client := pb.NewRecordShardClient(conn)
			curGsn := int64(0)
			for i := 0; i < 10; i++ {
				record := randString(10)
				res, err := client.Append(context.Background(), &pb.AppendRequest{
					Record: record,
					Color:  0,
				})
				require.NotNil(t, res)
				require.NoError(t, err)
				require.Equal(t, res.Record, record)
				require.True(t, curGsn <= res.Gsn, "GSN has to increase")
				checkC <- res.Gsn
				curGsn = res.Gsn
			}
			waitC <- true
		}()
	}

	for i := 0; i < threads; i++ {
		<-waitC
	}
	time.Sleep(time.Second)
	close(checkC)
}

// Tests if shard2 can read everything that shard1 has written
func TestWriteRead(t *testing.T) {
	writeSet := make(map[int64]string)
	fromGsn := int64(0)
	finishedWriting := make(chan bool)
	go func() {
		conn, err := grpc.Dial(shardIp1, grpc.WithInsecure())
		require.NoError(t, err)
		defer conn.Close()
		client := pb.NewRecordShardClient(conn)
		curGsn := int64(0)
		for i := 0; i < 10; i++ {
			record := randString(10)
			res, err := client.Append(context.Background(), &pb.AppendRequest{
				Record: record,
				Color:  0,
			})
			require.NotNil(t, res)
			require.NoError(t, err)
			require.Equal(t, res.Record, record)
			require.True(t, curGsn <= res.Gsn, "GSN has to increase")
			curGsn = res.Gsn
			if i == 0 {
				fromGsn = curGsn
			}
			writeSet[curGsn] = res.Record
		}
		finishedWriting <- true
	}()
	<-finishedWriting

	time.Sleep(time.Second)

	conn, err := grpc.Dial(shardIp2, grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()
	client := pb.NewRecordShardClient(conn)
	stream, err := client.Subscribe(context.Background(), &pb.ReadRequest{Gsn: fromGsn})
	for i := 0; i < 10; i++ {
		res, err := stream.Recv()
		require.NotNil(t, res)
		require.NoError(t, err)
		record, ok := writeSet[res.Gsn]
		require.True(t, ok)
		require.Equal(t, record, res.Record)
	}
}

func startRecordShard(ipAddr, parentIpAddr string) error {
	recShard, err := NewRecordShard(19, "tmp/test", time.Second)
	if err != nil {
		return err
	}
	err = recShard.Start(ipAddr, parentIpAddr)
	if err != nil {
		return err
	}
	return nil
}

func randString(n int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func checkDistinctGSN(t *testing.T, c chan int64) {
	gsnSet := make(map[int64]bool)
	for gsn := range c {
		_, ok := gsnSet[gsn]
		require.False(t, ok)
		gsnSet[gsn] = true
	}
}
