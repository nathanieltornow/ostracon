package recshard

import (
	"testing"
)

func TestBehaviour(t *testing.T) {
	// priWriteC, secWriteC, orderReqC chan *record, orderRespC chan *spb.OrderResponse
	//orderC := make(chan *record, 4)
	//cs, err := newColorService("tmp", 0, orderC)
	//require.NoError(t, err)
	//
	//go func() {
	//	err := cs.Start()
	//	require.NoError(t, err)
	//}()
	//
	//rec1 := record{r: "Hallo1", gsnWait: make(chan int64)}
	//rec2 := record{r: "Hallo2", gsnWait: make(chan int64)}
	//rec3 := record{r: "Hallo3", gsnWait: make(chan int64)}
	//rec4 := record{r: "Hallo4", gsnWait: make(chan int64)}
	//cs.priWriteC <- &rec1
	//cs.priWriteC <- &rec2
	//cs.priWriteC <- &rec3
	//cs.priWriteC <- &rec4
	//
	//time.Sleep(time.Second)
	//
	//cs.orderRespC <- &spb.OrderResponse{StartLsn: 0, StartGsn: 5, NumOfRecords: 4, Color: 0}
	//
	//gsn := <-rec1.gsnWait
	//require.Equal(t, gsn, int64(5))
	//gsn = <-rec2.gsnWait
	//require.Equal(t, gsn, int64(6))
	//gsn = <-rec3.gsnWait
	//require.Equal(t, gsn, int64(7))
	//gsn = <-rec4.gsnWait
	//require.Equal(t, gsn, int64(8))

}
