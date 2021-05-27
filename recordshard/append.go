package recordshard

import (
	"context"
	rpb "github.com/nathanieltornow/ostracon/recordshard/recordshardpb"
	"github.com/sirupsen/logrus"
)

func (rs *RecordShard) Append(ctx context.Context, request *rpb.AppendRequest) (*rpb.CommittedRecord, error) {
	if rs.parentConn == nil && rs.parentClient == nil {

		return nil, nil
	}

	// enqueue the new record in the write channel
	newRec := &record{record: request.Record, gsn: make(chan int64)}
	rs.writeC <- newRec
	gsn := <-newRec.gsn
	return &rpb.CommittedRecord{Gsn: gsn, Record: request.Record}, nil
}

func (rs *RecordShard) writeAppends() {
	for rec := range rs.writeC {
		rs.diskMu.Lock()
		lsn, err := rs.disk.Write(rec.record)
		rs.diskMu.Unlock()
		if err != nil {
			logrus.Fatalln(err)
		}
		rs.lsnToRecordMu.Lock()
		rs.lsnToRecord[lsn] = rec
		rs.lsnToRecordMu.Unlock()
		// update lsn to trigger orderRequests
		rs.curLsnMu.Lock()
		rs.curLsn = lsn
		rs.curLsnMu.Unlock()
	}
}
