package rshard

import (
	"context"
	pb "github.com/nathanieltornow/ostracon/rshard/rshardpb"
)

func (rs *RecordShard) Append(_ context.Context, request *pb.AppendRequest) (*pb.CommittedRecord, error) {
	rs.Lock()
	cs, ok := rs.colorToService[request.Color]
	rs.Unlock()
	if !ok {
		var err error
		cs, err = rs.addColor(request.Color)
		if err != nil {
			return nil, err
		}
	}
	rec := record{r: request.Record, color: request.Color, gsnWait: make(chan int64)}
	cs.priWriteC <- &rec

	gsn := <-rec.gsnWait
	return &pb.CommittedRecord{Record: rec.r, Gsn: gsn, Color: request.Color}, nil
}
