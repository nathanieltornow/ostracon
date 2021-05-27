package shard

import pb "github.com/nathanieltornow/ostracon/shard/shardpb"

func (s *Shard) SendOrderRequests(stream pb.Shard_GetOrderClient) {
	//lsnCopy := s.sn
	//for range time.Tick(s.batchingIntervall) {
	//	if lsnCopy == s.sn {
	//		continue
	//	}
	//	orderReq := pb.OrderRequest{StartLsn: lsnCopy, NumOfRecords: s.sn - lsnCopy}
	//	err := stream.Send(&orderReq)
	//
	//	lsnCopy = s.sn
	//	if err != nil {
	//		logrus.Fatalln("Failed to send order requests")
	//	}
	//
	//}
}

func (s *Shard) ReceiveOrderResponses(stream pb.Shard_GetOrderClient) {

	//for {
	//	in, err := stream.Recv()
	//	if err != nil {
	//		logrus.Fatalln("Failed to receive order requests")
	//	}
	//
	//	s.waitMapMu.Lock()
	//	for i := int64(0); i < in.NumOfRecords; i++ {
	//		if _, ok := s.waitMap[in.StartLsn+i]; ok {
	//			s.waitMap[in.StartLsn+i] <- in.StartGsn + i
	//		} else {
	//			logrus.Fatalln("Not in map")
	//		}
	//	}
	//	s.waitMapMu.Unlock()
	//}

}
