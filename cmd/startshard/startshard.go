package main

import (
	"flag"
	"github.com/nathanieltornow/ostracon/shard"
	"github.com/sirupsen/logrus"
	"time"
)

var (
	isRoot           = flag.Bool("isRoot", false, "Determines if started shard is root of the system")
	isSequencer      = flag.Bool("isSequencer", false, "Determines if started shard can be used as sequencer")
	ipAddr           = flag.String("ipAddr", "localhost:4000", "IP-Address of shard")
	parentIpAddr     = flag.String("parentIpAddr", "", "The IpAddress of the parent-shard")
	diskPath         = flag.String("diskPath", "tmp", "Path for the storage-files")
	batchingInterval = flag.Duration("interval", time.Second, "Intervall to request ordering")
)

func main() {
	flag.Parse()

	shardIpAddr := "localhost:4000"
	shard, err := shard.NewShard("tmp", true, true, time.Second)
	if err != nil {
		logrus.Fatalln("Failed creating shard")
	}
	logrus.Infof("Starting shard on %v", shardIpAddr)

	err = shard.Start(shardIpAddr, "")
	if err != nil {
		logrus.Fatalln("Failed starting shard")
	}
	//if *isRoot && *isSequencer {
	//
	//	s, err := shard.NewShard(*diskPath, true, true, *batchingInterval)
	//	if err != nil {
	//		logrus.Fatalln("Failed to create new shard", err)
	//	}
	//
	//	err = s.Start(*ipAddr, *parentIpAddr)
	//	if err != nil {
	//		logrus.Fatalln("Failed to start shard", err)
	//	}
	//
	//} else if !*isSequencer && !*isRoot {
	//
	//	s, err := shard.NewShard(*diskPath, false, false, *batchingInterval)
	//	if err != nil {
	//		logrus.Fatalln("Failed to create new shard", err)
	//	}
	//
	//	err = s.Start(*ipAddr, *parentIpAddr)
	//	if err != nil {
	//		logrus.Fatalln("Failed to start shard", err)
	//	}
	//
	//}
}
