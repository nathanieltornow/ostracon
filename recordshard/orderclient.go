package recordshard

import (
	spb "github.com/nathanieltornow/ostracon/shard/shardpb"
	"github.com/sirupsen/logrus"
	"time"
)

func (rs *RecordShard) sendOrderRequests(stream spb.Shard_GetOrderClient) {
	prevLsn := rs.curLsn
	for range time.Tick(rs.batchingInterval) {
		if rs.curLsn == prevLsn || rs.curLsn < 0 {
			continue
		}
		rs.curLsnMu.Lock()
		orderReq := spb.OrderRequest{StartLsn: prevLsn + 1, NumOfRecords: rs.curLsn - prevLsn}
		prevLsn = rs.curLsn
		rs.curLsnMu.Unlock()

		err := stream.Send(&orderReq)
		if err != nil {
			logrus.Fatalln("Failed to send order requests")
		}
	}
}

func (rs *RecordShard) receiveOrderResponses(stream spb.Shard_GetOrderClient) {

	for {
		in, err := stream.Recv()
		if err != nil {
			logrus.Fatalln("Failed to receive order requests")
		}
		err = rs.disk.Assign(0, in.StartLsn, int32(in.NumOfRecords), in.StartGsn)
		if err != nil {
			logrus.Fatalln("Failed to assign", err)
		}

		rs.waitCMapMu.Lock()
		for i := int64(0); i < in.NumOfRecords; i++ {
			rs.waitCMap[in.StartLsn+i] <- in.StartGsn + i
		}
		rs.waitCMapMu.Unlock()
	}
}
