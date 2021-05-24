package main

import (
	"github.com/nathanieltornow/ostracon/storage"
	"time"
)

func main() {
	disk, _ := storage.NewStorage("tmp")
	for range time.Tick(time.Second) {
		disk.Write("hallo")
	}
}
