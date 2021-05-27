package shard

import (
	pb "github.com/nathanieltornow/ostracon/shard/shardpb"
	"io"
)

func (s *Shard) GetOrder(stream pb.Shard_GetOrderServer) error {

	if s.isRoot {
		for {
			req, err := stream.Recv()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}

			s.snMu.Lock()
			res := pb.OrderResponse{StartGsn: s.sn, StartLsn: req.StartLsn, NumOfRecords: req.NumOfRecords}
			s.sn += req.NumOfRecords
			s.snMu.Unlock()
			if err := stream.Send(&res); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Shard) SendOrderResponses(stream pb.Shard_GetOrderServer) error {

	return nil
}
