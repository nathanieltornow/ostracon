package seqshard

import (
	"fmt"
	pb "github.com/nathanieltornow/ostracon/seqshard/seqshardpb"
	"github.com/sirupsen/logrus"
	"time"
)

func (s *SeqShard) sendOrderRequests(stream pb.Shard_GetOrderClient) {
	prevSns := make(map[int64]int64)
	counts := make(map[int64]int64)
	ticker := time.Tick(s.batchingIntervall)
	for {
		select {
		case <-ticker:
			for color, prevSn := range prevSns {
				count := counts[color]
				if count == 0 {
					continue
				}
				ordReq := pb.OrderRequest{StartLsn: prevSn, NumOfRecords: count, Color: color}
				fmt.Println("Sending", ordReq.String(), ordReq.Color)
				err := stream.Send(&ordReq)
				if err != nil {
					return
				}
				prevSns[color] += count
				counts[color] = 0
			}
		case iOR := <-s.orderReqsC:
			_, ok := prevSns[iOR.color]
			if !ok {
				prevSns[iOR.color] = 0
				counts[iOR.color] = 0
			}
			counts[iOR.color] += iOR.numOfRecords
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
			fmt.Println("getting", in.Color, in.StartLsn+i)
			pendOR := s.waitingOrderReqs[snColorTuple{color: in.Color, sn: in.StartLsn + i}]
			delete(s.waitingOrderReqs, snColorTuple{color: in.Color, sn: in.StartLsn + i})

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
