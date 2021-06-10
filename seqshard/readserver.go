package seqshard

import (
	pb "github.com/nathanieltornow/ostracon/seqshard/seqshardpb"
	"io"
)

func (s *Shard) ReportCommittedRecords(stream pb.Shard_ReportCommittedRecordsServer) error {

	c := make(chan *committedRecord, 4096)
	s.comRecStreamsMu.Lock()
	s.comRecStreams = append(s.comRecStreams, c)
	s.comRecStreamsMu.Unlock()
	go s.SendCommittedRecords(stream, c)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		s.incomingComRec <- &committedRecord{gsn: req.Gsn, record: req.Record}
	}

}

func (s *Shard) SendCommittedRecords(stream pb.Shard_ReportCommittedRecordsServer, c chan *committedRecord) {

	for rec := range c {
		err := stream.Send(&pb.CommittedRecord{Gsn: rec.gsn, Record: rec.record})
		if err != nil {
			return
		}
	}
}

func (s *Shard) broadcastComRec() {
	for rec := range s.incomingComRec {
		s.comRecStreamsMu.RLock()
		for _, c := range s.comRecStreams {
			c <- rec
		}
		s.comRecStreamsMu.RUnlock()
	}
}
