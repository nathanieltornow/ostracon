package rshard

import (
	"fmt"
)

func (rs *RecordShard) sendOrderRequests() error {
	stream := rs.orderStream
	if stream == nil {
		return fmt.Errorf("sendOrderRequests: order-stream is nil")
	}
	for orderReq := range rs.orderReqC {
		err := rs.orderStream.Send(orderReq)
		if err != nil {
			return err
		}
	}
	return nil
}

func (rs *RecordShard) receiveOrderResponses() error {
	stream := rs.orderStream
	if stream == nil {
		return fmt.Errorf("receiveOrderRequests: order-stream is nil")
	}
	for {
		in, err := stream.Recv()
		if err != nil {
			return err
		}
		rs.Lock()
		cs, ok := rs.colorToService[in.Color]
		rs.Unlock()
		if !ok {
			return fmt.Errorf("receiveOrderRequests: Unknown color")
		}
		cs.orderRespC <- in
	}
}
