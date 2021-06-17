package recordshard

import (
	rpb "github.com/nathanieltornow/ostracon/recordshard/recordshardpb"
	spb "github.com/nathanieltornow/ostracon/seqshard/seqshardpb"
	"github.com/sirupsen/logrus"
)

func (rs *RecordShard) Subscribe(r *rpb.ReadRequest, stream rpb.RecordShard_SubscribeServer) error {
	sC := make(chan *rpb.CommittedRecord, 4096)

	go rs.readStoredReplicas(r.Gsn, sC)
	rs.subscriberCsMu.Lock()
	rs.subscribeCs = append(rs.subscribeCs, sC)
	rs.subscriberCsMu.Unlock()
	for rec := range sC {
		// send back to client
		err := stream.Send(rec)
		if err != nil {
			return err
		}
	}
	return nil
}

func (rs *RecordShard) receiveCommittedRecords(stream spb.Shard_ReportCommittedRecordsClient) {
	for {
		in, err := stream.Recv()
		rs.replicaC <- &rpb.CommittedRecord{Record: in.Record, Gsn: in.Gsn}
		if err != nil {
			logrus.Fatalln("Failed to receive committed records")
		}
		rs.subscriberCsMu.RLock()
		for _, c := range rs.subscribeCs {
			c <- &rpb.CommittedRecord{Record: in.Record, Gsn: in.Gsn}
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

func (rs *RecordShard) readStoredReplicas(fromGsn int64, c chan *rpb.CommittedRecord) {
	//i := fromGsn
	//to := rs.disk.GetCurrentLsn()
	//for ; i <= to; i++ {
	//	rec, err := rs.disk.ReadGSN(1, i)
	//	if err != nil {
	//		continue
	//	}
	//	c <- &rpb.CommittedRecord{Record: rec, Gsn: i}
	//}
}

func (rs *RecordShard) storeReplicas() {
	//for r := range rs.replicaC {
	//	lsn, _ := rs.disk.WriteToPartition(1, r.Record)
	//	_ = rs.disk.Assign(1, lsn, 1, r.Gsn)
	//}
}
