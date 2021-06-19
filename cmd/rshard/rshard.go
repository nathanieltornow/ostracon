package main

import (
	"github.com/nathanieltornow/ostracon/rshard"
	"github.com/sirupsen/logrus"
	"time"
)

func main() {
	rs, err := rshard.NewRecordShard("tmp2", time.Microsecond*100)
	if err != nil {
		logrus.Fatalln(err)
	}
	err = rs.Start("localhost:6000", "localhost:4000")
	if err != nil {
		logrus.Fatalln(err)
	}
}
