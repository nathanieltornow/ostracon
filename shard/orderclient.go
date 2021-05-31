package shard

import (
	pb "github.com/nathanieltornow/ostracon/shard/shardpb"
	"github.com/sirupsen/logrus"
	"time"
)

func (s *Shard) SendOrderRequests(stream pb.Shard_GetOrderClient) {
	timeC := time.Tick(s.batchingIntervall * 5)
	count := int64(0)
	prevSn := int64(0)
	for {
		select {
		case <-timeC:
			if count == 0 {
				continue
			}
			ordReq := pb.OrderRequest{StartLsn: prevSn, NumOfRecords: count}

			s.snMu.Lock()
			prevSn = s.sn
			s.snMu.Unlock()

			err := stream.Send(&ordReq)
			if err != nil {
				return
			}
			count = 0

		case iOR := <-s.incomingOR:
			count += iOR.numOfRecords
		}
	}
}

func (s *Shard) ReceiveOrderResponses(stream pb.Shard_GetOrderClient) {

	for {
		in, err := stream.Recv()
		if err != nil {
			logrus.Fatalln("Failed to receive order requests")
		}

		s.snMu.Lock()
		for i := int64(0); i < in.NumOfRecords; {
			pendOR := s.snToPendingOR[in.StartLsn+i]
			s.streamToOR[pendOR.stream] <- &orderResponse{
				numOfRecords: pendOR.numOfRecords,
				startLsn:     pendOR.startLsn,
				startGsn:     in.StartGsn + i,
			}
			i += pendOR.numOfRecords
		}
		s.snMu.Unlock()
	}

}
