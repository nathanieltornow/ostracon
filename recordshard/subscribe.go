package recordshard

import (
	rpb "github.com/nathanieltornow/ostracon/recordshard/recordshardpb"
	spb "github.com/nathanieltornow/ostracon/seqshard/seqshardpb"
	"github.com/sirupsen/logrus"
)

func (rs *RecordShard) Subscribe(r *rpb.ReadRequest, stream rpb.RecordShard_SubscribeServer) error {
	sC := make(chan *rpb.CommittedRecord, 4096)

	go rs.readStoredReplicas(r.Gsn, r.Color, sC)
	rs.subscriberCsMu.Lock()
	rs.subscribeCs = append(rs.subscribeCs, sC)
	rs.subscriberCsMu.Unlock()
	for rec := range sC {
		// send back to client
		if rec.Color == r.Color {
			err := stream.Send(rec)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (rs *RecordShard) receiveCommittedRecords(stream spb.Shard_ReportCommittedRecordsClient) {
	for {
		in, err := stream.Recv()
		rs.replicaC <- &rpb.CommittedRecord{Record: in.Record, Gsn: in.Gsn, Color: in.Color}
		if err != nil {
			logrus.Fatalln("Failed to receive committed records")
		}
		rs.subscriberCsMu.RLock()
		for _, c := range rs.subscribeCs {
			c <- &rpb.CommittedRecord{Record: in.Record, Gsn: in.Gsn, Color: in.Color}
		}
		rs.subscriberCsMu.RUnlock()
	}
}

func (rs *RecordShard) sendCommittedRecords(stream spb.Shard_ReportCommittedRecordsClient) {
	for rec := range rs.newComRecC {
		err := stream.Send(rec)
		if err != nil {
			logrus.Fatalln("Failed to send committed records")
		}
	}
}

func (rs *RecordShard) readStoredReplicas(fromGsn, color int64, c chan *rpb.CommittedRecord) {
	i := fromGsn
	to, _ := rs.disk.GetCurrentLsn(color, false)
	for ; i <= to; i++ {
		rec, err := rs.disk.Read(i, color)
		if err != nil {
			continue
		}
		c <- &rpb.CommittedRecord{Record: rec, Gsn: i, Color: color}
	}
}

func (rs *RecordShard) storeReplicas() {
	for r := range rs.replicaC {
		lsn, _ := rs.disk.Write(false, r.Record, r.Color)
		_ = rs.disk.Assign(false, lsn, r.Gsn, r.Color, 1)
	}
}
