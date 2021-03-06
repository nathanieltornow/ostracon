package main

import (
	"flag"
	"github.com/nathanieltornow/ostracon/recshard"
	"github.com/nathanieltornow/ostracon/seqshard"
	"github.com/sirupsen/logrus"
	"time"
)

var (
	rec              = flag.Bool("rec", false, "SeqShard will be a record-shard")
	storagePath      = flag.String("storagePath", "tmp", "Path to storage directory")
	isRoot           = flag.Bool("isRoot", false, "Determines if started seqshard is root of the system")
	ipAddr           = flag.String("ipAddr", ":4000", "IP-Address of seqshard")
	parentIpAddr     = flag.String("parentIpAddr", "", "The IpAddress of the parent-seqshard")
	batchingInterval = flag.Duration("interval", time.Second, "Intervall to request ordering")
	color            = flag.Int64("color", -1, "color of the shard")
)

func main() {
	flag.Parse()

	if *rec {
		err := startRecordShard()
		if err != nil {
			logrus.Fatalf("Error starting recordshard %v", err)
		}
	}
	err := startSeqShard()
	if err != nil {
		logrus.Fatalf("Error starting sequencershard %v", err)
	}

}

func startRecordShard() error {
	recShard, err := recshard.NewRecordShard(*storagePath, *batchingInterval)
	if err != nil {
		return err
	}
	err = recShard.Start(*ipAddr, *parentIpAddr)
	if err != nil {
		return err
	}
	return nil
}

func startSeqShard() error {
	s, err := seqshard.NewSeqShard(*color, *isRoot, *batchingInterval)
	if err != nil {
		return err
	}
	err = s.Start(*ipAddr, *parentIpAddr)
	if err != nil {
		return err
	}
	return nil
}
