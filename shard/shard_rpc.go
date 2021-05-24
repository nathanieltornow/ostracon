package main

import (
	"context"
	pb "github.com/nathanieltornow/ostracon/shard/shardpb"
	"io"
)

func (s *Shard) Append(ctx context.Context, request *pb.AppendRequest) (*pb.CommittedRecord, error) {

	return nil, nil
}

func (s *Shard) GetOrder(stream pb.Shard_GetOrderServer) error {

	if s.isRoot && s.isSequencer {

		for {
			req, err := stream.Recv()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}

			s.lsnMu.Lock()
			res := pb.OrderResponse{StartGsn: s.lsn, StartLsn: req.StartLsn}
			s.lsn += req.NumOfRecords
			s.lsnMu.Unlock()

			if err := stream.Send(&res); err != nil {
				return err
			}
		}
	}

	return nil
}
