package seqshard

import (
	pb "github.com/nathanieltornow/ostracon/seqshard/seqshardpb"
	"github.com/sirupsen/logrus"
	"time"
)

func (s *SeqShard) sendOrderRequests(stream pb.Shard_GetOrderClient) {
	count := int64(0)
	prevSn := int64(0)
	ticker := time.Tick(s.batchingIntervall)
	for {
		select {
		case <-ticker:
			if count == 0 {
				continue
			}
			ordReq := pb.OrderRequest{StartLsn: prevSn, NumOfRecords: count}

			prevSn += count

			err := stream.Send(&ordReq)
			if err != nil {
				return
			}
			count = 0

		case iOR := <-s.orderReqsC:
			count += iOR.numOfRecords
		}
	}
}

func (s *SeqShard) receiveOrderResponses(stream pb.Shard_GetOrderClient) {

	for {
		in, err := stream.Recv()
		if err != nil {
			logrus.Fatalln("Failed to receive order requests")
		}

		s.snMu.Lock()
		for i := int64(0); i < in.NumOfRecords; {
			pendOR := s.waitingOrderReqs[in.StartLsn+i]
			delete(s.waitingOrderReqs, in.StartLsn+i)

			s.orderRespCs[pendOR.stream] <- &orderResponse{
				numOfRecords: pendOR.numOfRecords,
				startLsn:     pendOR.startLsn,
				startGsn:     in.StartGsn + i,
				color:        pendOR.color,
			}
			i += pendOR.numOfRecords
		}
		s.snMu.Unlock()
	}

}
