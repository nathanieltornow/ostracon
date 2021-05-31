package main

import (
	"flag"
	"github.com/nathanieltornow/ostracon/shard"
	"github.com/sirupsen/logrus"
	"time"
)

var (
	isRoot           = flag.Bool("isRoot", false, "Determines if started shard is root of the system")
	ipAddr           = flag.String("ipAddr", "localhost:4000", "IP-Address of shard")
	parentIpAddr     = flag.String("parentIpAddr", "", "The IpAddress of the parent-shard")
	diskPath         = flag.String("diskPath", "tmp", "Path for the storage-files")
	batchingInterval = flag.Duration("interval", time.Second, "Intervall to request ordering")
)

func main() {
	flag.Parse()

	shardIpAddr := *ipAddr
	s, err := shard.NewShard("tmp", *isRoot, *batchingInterval)
	if err != nil {
		logrus.Fatalln("Failed creating s")
	}
	logrus.Infof("Starting s on %v", shardIpAddr)

	err = s.Start(shardIpAddr, *parentIpAddr)
	if err != nil {
		logrus.Fatalln("Failed starting s:", err)
	}
}
