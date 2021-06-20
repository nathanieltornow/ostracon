package recshard

import (
	"fmt"
	pb "github.com/nathanieltornow/ostracon/recshard/recshardpb"
	spb "github.com/nathanieltornow/ostracon/seqshard/seqshardpb"
)

func (rs *RecordShard) Subscribe(r *pb.ReadRequest, stream pb.RecordShard_SubscribeServer) error {
	rs.Lock()
	cs, ok := rs.colorToService[r.Color]
	rs.Unlock()
	if !ok {
		return fmt.Errorf("color isn't supported on this shard")
	}
	subscribeC := make(chan *pb.CommittedRecord, 2048)
	fmt.Println("Start subs")
	subId := cs.subscribe(r.Gsn, subscribeC)
	for comRec := range subscribeC {
		err := stream.Send(comRec)
		if err != nil {
			cs.removeSubscription(subId)
			return err
		}
	}
	return nil
}

func (rs *RecordShard) sendCommittedRecords() error {
	for comRec := range rs.comRecC {
		comR := spb.CommittedRecord{Color: comRec.Color, Record: comRec.Record, Gsn: comRec.Gsn}
		err := rs.subsStream.Send(&comR)
		if err != nil {
			return nil
		}
	}
	return nil
}

func (rs *RecordShard) receiveCommittedRecords() error {
	for {
		in, err := rs.subsStream.Recv()
		if err != nil {
			return err
		}
		rs.Lock()
		cs, ok := rs.colorToService[in.Color]
		rs.Unlock()
		if !ok {
			cs, err = rs.addColor(in.Color)
			if err != nil {
				return err
			}
		}
		cs.secWriteC <- in
	}
}
