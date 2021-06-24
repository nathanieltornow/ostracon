package main

import (
	"flag"
	"fmt"
	"github.com/eiannone/keyboard"
	"github.com/nathanieltornow/ostracon/client"
	"github.com/sirupsen/logrus"
	"os"
)

var (
	subscribeFlag = flag.Bool("s", false, "Start subscription")
	record        = flag.String("record", "Hallo", "Record to append")
	appendFlag    = flag.Bool("a", false, "Append record")
	gsn           = flag.Int64("gsn", 0, "Global sequence number")
	color         = flag.Int64("color", 0, "Color to append/ subscribe to")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\nAppending: ./ostracon -a -color <int> -record <string>\n\n"+
			"Subscribing: ./ostracon -s -color <int> -gsn <int>\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if !*subscribeFlag && !*appendFlag {
		flag.Usage()
		os.Exit(1)
	}

	if *subscribeFlag {
		fmt.Printf("Subscribing to color %v. Press 'q' to stop.\n", *color)
		resultC := make(chan *client.Record, 64)
		err := client.Subscribe(*gsn, *color, resultC)
		if err != nil {
			logrus.Errorln(err)
			return
		}
		go func() {
			for {
				char, _, err := keyboard.GetSingleKey()
				if err != nil {
					panic(err)
				}
				if char == 'q' {
					close(resultC)
				}
			}
		}()

		for rec := range resultC {
			fmt.Printf("Record: \"%v\" | Sequence number: %v | Color: %v\n", rec.Record, rec.Gsn, rec.Color)
		}
	} else if *appendFlag {

		rec, err := client.Append(*color, *record)
		if err != nil {
			logrus.Errorln(err)
			return
		}
		fmt.Printf("Appended record \"%v\" with sequence number %v to color %v\n", rec.Record, rec.Gsn, rec.Color)
	}
}
