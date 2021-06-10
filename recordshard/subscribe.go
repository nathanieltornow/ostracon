package recordshard

import (
	rpb "github.com/nathanieltornow/ostracon/recordshard/recordshardpb"
	spb "github.com/nathanieltornow/ostracon/seqshard/seqshardpb"
	"github.com/sirupsen/logrus"
)

func (rs *RecordShard) Subscribe(_ *rpb.Empty, stream rpb.RecordShard_SubscribeServer) error {
	go rs.receiveCommittedRecords(*rs.parentReadStream)

	for rec := range rs.subscribeC {
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
		if err != nil {
			logrus.Fatalln("Failed to receive committed records")
		}
		rs.subscribeC <- &rpb.CommittedRecord{Record: in.Record, Gsn: in.Gsn}
	}
}

func (rs *RecordShard) sendCommittedRecords(stream spb.Shard_ReportCommittedRecordsClient) {
	for rec := range rs.newComRecC {
		err := stream.Send(rec)
		if err != nil {
			logrus.Fatalln("Failed to send committed reords")
		}
	}
}
