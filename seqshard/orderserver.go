package seqshard

import (
	"fmt"
	pb "github.com/nathanieltornow/ostracon/seqshard/seqshardpb"
	"io"
)

func (s *SeqShard) GetOrder(stream pb.Shard_GetOrderServer) error {
	s.orderRespCsMu.Lock()
	s.orderRespCs[stream] = make(chan *orderResponse, 4096)
	s.orderRespCsMu.Unlock()

	go s.sendOrderResponses(stream)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		if s.isRoot || req.Color == s.color {
			s.snMu.Lock()
			res := pb.OrderResponse{StartGsn: s.sn, StartLsn: req.StartLsn, NumOfRecords: req.NumOfRecords, Color: req.Color}
			s.sn += req.NumOfRecords
			s.snMu.Unlock()
			fmt.Println("Sending", res.String())
			if err := stream.Send(&res); err != nil {
				return err
			}

		} else {

			s.snMu.Lock()
			oR := orderRequest{stream: stream, numOfRecords: req.NumOfRecords, startLsn: req.StartLsn, color: req.Color}
			s.waitingOrderReqs[snColorTuple{color: req.Color, sn: req.StartLsn}] = &oR
			s.snMu.Unlock()
			s.orderReqsC <- &oR
		}
	}
}

// sendOrderResponses handles a specific order-stream to send back all finished order-requests in its channel
func (s *SeqShard) sendOrderResponses(stream pb.Shard_GetOrderServer) {
	for finishedOR := range s.orderRespCs[stream] {
		res := pb.OrderResponse{StartLsn: finishedOR.startLsn, StartGsn: finishedOR.startGsn, NumOfRecords: finishedOR.numOfRecords, Color: finishedOR.color}
		if err := stream.Send(&res); err != nil {
			return
		}
	}
}
