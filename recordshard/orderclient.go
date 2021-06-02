package recordshard

import (
	spb "github.com/nathanieltornow/ostracon/shard/shardpb"
	"github.com/sirupsen/logrus"
	"log"
)

func (rs *RecordShard) sendOrderRequests(stream spb.Shard_GetOrderClient) {
	prevLsn := rs.curLsn
	for {
		rs.curLsnMu.Lock()
		curLsnCop := rs.curLsn
		rs.curLsnMu.Unlock()

		if curLsnCop == prevLsn || curLsnCop < 0 {
			continue
		}
		if curLsnCop < prevLsn {

			log.Fatalln("WTF", curLsnCop, prevLsn)
		}
		orderReq := spb.OrderRequest{StartLsn: prevLsn + 1, NumOfRecords: curLsnCop - prevLsn}
		prevLsn = curLsnCop

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

		// assign gsn for subsequent entries
		rs.diskMu.Lock()
		err = rs.disk.Assign(0, in.StartLsn, int32(in.NumOfRecords), in.StartGsn)
		rs.diskMu.Unlock()
		if err != nil {
			logrus.Fatalln("Failed to assign", err)
		}

		rs.lsnToRecordMu.Lock()
		for i := int64(0); i < in.NumOfRecords; i++ {
			rs.lsnToRecord[in.StartLsn+i].gsn <- in.StartGsn + i
			rs.newComRecC <- &spb.CommittedRecord{Record: rs.lsnToRecord[in.StartLsn+i].record, Gsn: in.StartGsn + i}
			delete(rs.lsnToRecord, in.StartLsn+i)
		}
		rs.lsnToRecordMu.Unlock()
	}
}
