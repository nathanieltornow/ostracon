package seqshard

import (
	pb "github.com/nathanieltornow/ostracon/seqshard/seqshardpb"
	"io"
)

func (s *SeqShard) ReportCommittedRecords(stream pb.Shard_ReportCommittedRecordsServer) error {

	c := make(chan *committedRecord, 4096)
	s.comRecCsOutgoingMu.Lock()
	s.comRecCsOutgoing = append(s.comRecCsOutgoing, c)
	s.comRecCsOutgoingMu.Unlock()
	go s.sendCommittedRecords(stream, c)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		s.comRecCIncoming <- &committedRecord{gsn: req.Gsn, record: req.Record, color: req.Color}
	}

}

func (s *SeqShard) sendCommittedRecords(stream pb.Shard_ReportCommittedRecordsServer, c chan *committedRecord) {
	for rec := range c {
		err := stream.Send(&pb.CommittedRecord{Gsn: rec.gsn, Record: rec.record, Color: rec.color})
		if err != nil {
			return
		}
	}
}

func (s *SeqShard) broadcastComRec() {
	for rec := range s.comRecCIncoming {
		s.comRecCsOutgoingMu.RLock()
		for _, c := range s.comRecCsOutgoing {
			c <- rec
		}
		s.comRecCsOutgoingMu.RUnlock()
	}
}
