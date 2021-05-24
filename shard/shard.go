package main

import (
	pb "github.com/nathanieltornow/ostracon/shard/shardpb"
	"github.com/nathanieltornow/ostracon/storage"
	"time"
)

type Shard struct {
	pb.UnimplementedShardServer
	disk              *storage.Storage
	parentClient      *pb.ShardClient
	lsn               int64
	batchingIntervall time.Duration
}

func (s *Shard) SendOrderRequests() error {
	return nil
}

func (s *Shard) ReceiveOrderResponses() error {
	return nil
}
