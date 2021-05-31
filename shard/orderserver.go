package shard

import (
	pb "github.com/nathanieltornow/ostracon/shard/shardpb"
	"io"
)

func (s *Shard) GetOrder(stream pb.Shard_GetOrderServer) error {
	if s.isRoot {
		s.snMu.Lock()
		s.sn = 0
		s.snMu.Unlock()
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

	s.streamToORMu.Lock()
	s.streamToOR[stream] = make(chan *orderResponse, 4096)
	s.streamToORMu.Unlock()

	go s.SendOrderResponses(stream)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		s.snMu.Lock()
		oR := orderRequest{stream: stream, numOfRecords: req.NumOfRecords, startLsn: req.StartLsn}
		s.snToPendingOR[s.sn] = &oR
		s.incomingOR <- &oR
		s.sn += req.NumOfRecords
		s.snMu.Unlock()
	}

}

// SendOrderResponses handles a specific order-stream to send back all finished order-requests in its channel
func (s *Shard) SendOrderResponses(stream pb.Shard_GetOrderServer) error {
	for finishedOR := range s.streamToOR[stream] {
		res := pb.OrderResponse{StartLsn: finishedOR.startLsn, StartGsn: finishedOR.startGsn, NumOfRecords: finishedOR.numOfRecords}
		if err := stream.Send(&res); err != nil {
			return err
		}
	}
	return nil
}
