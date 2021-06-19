package main

import (
	"flag"
	"github.com/nathanieltornow/ostracon/rshard"
	"github.com/sirupsen/logrus"
	"time"
)

var (
	ipAddr       = flag.String("ipAddr", ":4000", "IP-Address of seqshard")
	parentIpAddr = flag.String("parentIpAddr", "", "The IpAddress of the parent-seqshard")
)

func main() {
	flag.Parse()

	rs, err := rshard.NewRecordShard("tmp2", time.Microsecond*100)
	if err != nil {
		logrus.Fatalln(err)
	}
	err = rs.Start(*ipAddr, *parentIpAddr)
	if err != nil {
		logrus.Fatalln(err)
	}
}
