package recordshard

import (
	spb "github.com/nathanieltornow/ostracon/seqshard/seqshardpb"
	"github.com/sirupsen/logrus"
	"time"
)

func (rs *RecordShard) sendOrderRequests(stream spb.Shard_GetOrderClient) {
	for range time.Tick(rs.batchingInterval) {
		rs.curLsnMu.Lock()
		for color, prevLsn := range rs.colorToPrevLsn {
			rs.colorToLsnMu.Lock()
			lsn, _ := rs.colorToLsn[color]
			rs.colorToLsnMu.Lock()
			if lsn == prevLsn {
				continue
			}
			orderReq := spb.OrderRequest{StartLsn: prevLsn + 1, NumOfRecords: lsn - prevLsn, Color: color}
			err := stream.Send(&orderReq)
			if err != nil {
				logrus.Fatalln("Failed to send order requests")
			}
			rs.colorToPrevLsn[color] = lsn
		}
		rs.curLsnMu.Unlock()

	}
}

func (rs *RecordShard) receiveOrderResponses(stream spb.Shard_GetOrderClient) {

	for {
		in, err := stream.Recv()
		if err != nil {
			logrus.Fatalln("Failed to receive order requests")
		}

		// assign gsn for subsequent entries
		err = rs.disk.Assign(true, in.StartLsn, in.StartGsn, in.Color, in.NumOfRecords)
		if err != nil {
			logrus.Fatalln("Failed to assign", err)
		}

		rs.lsnToRecordMu.Lock()
		for i := int64(0); i < in.NumOfRecords; i++ {
			rs.lsnToRecord[lsnColorTuple{color: in.Color, lsn: in.StartLsn + i}].gsn <- in.StartGsn + i
			rs.newComRecC <- &spb.CommittedRecord{
				Record: rs.lsnToRecord[lsnColorTuple{color: in.Color, lsn: in.StartLsn + i}].record,
				Gsn:    in.StartGsn + i, Color: in.Color,
			}
			delete(rs.lsnToRecord, lsnColorTuple{color: in.Color, lsn: in.StartLsn + i})
		}
		rs.lsnToRecordMu.Unlock()
	}
}
