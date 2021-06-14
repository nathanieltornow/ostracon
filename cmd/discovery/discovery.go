package main

import (
	"github.com/nathanieltornow/ostracon/discovery"
	"github.com/sirupsen/logrus"
)

func main() {
	dS := discovery.NewDiscoveryServer()
	err := dS.Start(":9999")
	if err != nil {
		logrus.Fatalf("Error starting server %v", err)
	}
}
