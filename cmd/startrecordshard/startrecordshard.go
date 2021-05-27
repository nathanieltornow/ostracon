package main

import (
	"github.com/nathanieltornow/ostracon/recordshard"
	"github.com/sirupsen/logrus"
	"time"
)

func main() {
	recShard, err := recordshard.NewRecordShard("tmp", time.Millisecond)
	if err != nil {
		logrus.Fatalln("Failed creating shard")
	}

	err = recShard.Start("localhost:6000", "localhost:4000")
	if err != nil {
		logrus.Fatalln("Failed starting shard")
	}
}
