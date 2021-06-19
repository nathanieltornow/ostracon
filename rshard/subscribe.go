package rshard

import (
	pb "github.com/nathanieltornow/ostracon/rshard/rshardpb"
)

func (rs *RecordShard) Subscribe(r *pb.ReadRequest, stream pb.RecordShard_SubscribeServer) error {
	return nil
}
