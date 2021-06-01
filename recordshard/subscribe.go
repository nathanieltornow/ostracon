package recordshard

import rpb "github.com/nathanieltornow/ostracon/recordshard/recordshardpb"

func (rs *RecordShard) Subscribe(_ *rpb.Empty, stream rpb.RecordShard_SubscribeServer) error {

	return nil
}
