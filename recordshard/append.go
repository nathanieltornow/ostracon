package recordshard

import (
	"context"
	rpb "github.com/nathanieltornow/ostracon/recordshard/recordshardpb"
	"github.com/sirupsen/logrus"
)

func (rs *RecordShard) Append(ctx context.Context, request *rpb.AppendRequest) (*rpb.CommittedRecord, error) {
	if rs.parentConn == nil && rs.parentClient == nil {
		lsn, err := rs.disk.Write(request.Record)
		if err != nil {
			return nil, err
		}
		err = rs.disk.Assign(0, lsn, 1, lsn)
		return &rpb.CommittedRecord{Gsn: lsn, Record: request.Record}, nil
	}

	// enqueue the new record in the write channel
	newRec := &record{record: request.Record, lsn: make(chan int64)}
	rs.writeC <- newRec

	// wait for lsn, which is assigned after the record was written
	lsn := <-newRec.lsn

	// put lsn in waitMap to wait for assigned gsn
	gsnC := make(chan int64)
	rs.waitCMapMu.Lock()
	rs.waitCMap[lsn] = gsnC
	rs.waitCMapMu.Unlock()

	// update lsn to make trigger orderRequests
	rs.curLsnMu.Lock()
	rs.curLsn = lsn
	rs.curLsnMu.Unlock()

	// wait for gsn to be assigned
	gsn := <-rs.waitCMap[lsn]

	rs.waitCMapMu.Lock()
	delete(rs.waitCMap, lsn)
	rs.waitCMapMu.Unlock()

	return &rpb.CommittedRecord{Gsn: gsn, Record: request.Record}, nil
}

func (rs *RecordShard) writeAppends() {
	for rec := range rs.writeC {
		lsn, err := rs.disk.Write(rec.record)
		if err != nil {
			logrus.Fatalln(err)
		}
		rec.lsn <- lsn
	}
}
