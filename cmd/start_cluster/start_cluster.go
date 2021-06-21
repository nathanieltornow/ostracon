package main

import (
	"github.com/nathanieltornow/ostracon/recshard"
	"github.com/nathanieltornow/ostracon/seqshard"
	"github.com/nathanieltornow/ostracon/util"
	"github.com/sirupsen/logrus"
	"time"
)

func main() {
	config, err := util.ParseConfig()
	if err != nil {
		logrus.Fatalln("failed to parse config-file: %v", err)
	}
	waitC := make(chan bool)

	var retErr error
	for _, shard := range config.Shards {
		time.Sleep(time.Second)
		if shard.T == "sequencer" {
			go startSeqShard(shard.IP, shard.ParentIP, shard.Color, shard.Root, shard.Interval)
		} else if shard.T == "record" {
			go startRecordShard(shard.IP, shard.ParentIP, shard.Disk, shard.Interval)
		}

		if retErr != nil {
			logrus.Fatalln("failed to start shard: %v", retErr)
		}
	}
	<-waitC

}

func startRecordShard(ipAddr, parentIpAddr, storagePath string, batchingInterval time.Duration) error {
	recShard, err := recshard.NewRecordShard(storagePath, batchingInterval)
	if err != nil {
		return err
	}
	err = recShard.Start(ipAddr, parentIpAddr)
	if err != nil {
		return err
	}
	return nil
}

func startSeqShard(ipAddr, parentIpAddr string, color int64, isRoot bool, batchingInterval time.Duration) error {
	s, err := seqshard.NewSeqShard(color, isRoot, batchingInterval)
	if err != nil {
		return err
	}
	err = s.Start(ipAddr, parentIpAddr)
	if err != nil {
		return err
	}
	return nil
}
