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

	// enqueue a new record in the write channel
	newRec := &record{record: request.Record, gsn: make(chan int64)}
	rs.writeC <- newRec

	// wait for the gsn to be assigned
	gsn := <-newRec.gsn
	return &rpb.CommittedRecord{Gsn: gsn, Record: request.Record}, nil
}

func (rs *RecordShard) writeAppends() {
	// consumes the write channel to write all incoming records
	for rec := range rs.writeC {
		// write the next record
		lsn, err := rs.disk.Write(rec.record)
		if err != nil {
			logrus.Fatalln(err)
		}
		// add record to map, so the gsn can be assigned for given lsn
		rs.lsnToRecordMu.Lock()
		rs.lsnToRecord[lsn] = rec
		rs.lsnToRecordMu.Unlock()
		// update lsn to trigger orderRequests
		rs.curLsnMu.Lock()
		rs.curLsn = lsn
		rs.curLsnMu.Unlock()
	}
}
