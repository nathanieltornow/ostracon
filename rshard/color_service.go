package rshard

import (
	"fmt"
	pb "github.com/nathanieltornow/ostracon/rshard/rshardpb"
	"github.com/nathanieltornow/ostracon/rshard/storage"
	spb "github.com/nathanieltornow/ostracon/seqshard/seqshardpb"
	"sync"
	"time"
)

type record struct {
	lsn     int64
	r       string
	color   int64
	gsnWait chan int64
}

type colorService struct {
	sync.Mutex
	disk *storage.Storage

	color int64

	interval time.Duration

	orderC chan *record

	// Records that have to be stored by this color will be inserted in priWriteC
	priWriteC chan *record
	secWriteC chan *spb.CommittedRecord

	comRecC chan *spb.CommittedRecord

	secGsn   int64
	secGsnMu sync.Mutex

	// stored records that are waiting for a gsnWait will be inserted orderReqC
	orderReqC  chan *spb.OrderRequest
	orderRespC chan *spb.OrderResponse

	subCsCnt int64
	subCs    map[int64]chan *pb.CommittedRecord
	subCsMu  sync.Mutex

	lsnToRecord map[int64]*record
}

func newColorService(storagePath string, color int64, orderReqC chan *spb.OrderRequest,
	comRecC chan *spb.CommittedRecord, interval time.Duration) (*colorService, error) {

	disk, err := storage.NewStorage(storagePath, 0, 2, 10000000)
	if err != nil {
		return nil, err
	}
	cs := new(colorService)
	cs.disk = disk
	cs.color = color
	cs.interval = interval
	cs.priWriteC = make(chan *record, 2048)
	cs.secWriteC = make(chan *spb.CommittedRecord, 2048)
	cs.orderRespC = make(chan *spb.OrderResponse, 2048)
	cs.orderReqC = orderReqC
	cs.orderC = make(chan *record, 2048)
	cs.subCs = make(map[int64]chan *pb.CommittedRecord, 0)
	cs.lsnToRecord = make(map[int64]*record)
	cs.comRecC = comRecC
	return cs, nil
}

func (c *colorService) Start() error {
	var retErr error
	waitC := make(chan bool)
	go func() {
		err := c.primaryWrites()
		if err != nil {
			retErr = err
			waitC <- true
		}
	}()
	go func() {
		err := c.secondaryWrites()
		if err != nil {
			retErr = err
			waitC <- true
		}
	}()
	go func() {
		err := c.handleOrderResponses()
		if err != nil {
			retErr = err
			waitC <- true
		}
	}()
	go func() {
		err := c.createOrderRequests()
		if err != nil {
			retErr = err
			waitC <- true
		}
	}()
	<-waitC
	return retErr
}

// WORKER-FUNCTIONS

func (c *colorService) primaryWrites() error {
	for rec := range c.priWriteC {
		if rec.color != c.color {
			return fmt.Errorf("got record for wrong color")
		}
		lsn, err := c.disk.Write(rec.r)
		if err != nil {
			return fmt.Errorf("failed to write record: %v", rec.r)
		}
		rec.lsn = lsn
		c.Lock()
		c.lsnToRecord[lsn] = rec
		c.Unlock()
		c.orderC <- rec

	}
	return nil
}

func (c *colorService) secondaryWrites() error {
	for rec := range c.secWriteC {
		if rec.Color != c.color {
			return fmt.Errorf("got record for wrong color")
		}
		lsn, err := c.disk.WriteToPartition(1, rec.Record)
		if err != nil {
			return fmt.Errorf("failed to sec-write record: %v", rec.Record)
		}
		err = c.disk.Assign(1, lsn, 1, rec.Gsn)
		if err != nil {
			return fmt.Errorf("failed to sec-assign record: %v", rec.Record)
		}
		c.secGsnMu.Lock()
		c.secGsn++
		c.secGsnMu.Unlock()
		c.subCsMu.Lock()
		for _, comRecC := range c.subCs {
			comRecC <- &pb.CommittedRecord{Color: c.color, Gsn: rec.Gsn, Record: rec.Record}
		}
		c.subCsMu.Unlock()
	}
	return nil
}

func (c *colorService) createOrderRequests() error {
	ticker := time.Tick(c.interval)
	startSn := int64(0)
	count := int64(0)
	b := true
	for {
		select {

		case <-ticker:
			if count == 0 {
				continue
			}
			orderReq := spb.OrderRequest{StartLsn: startSn, NumOfRecords: count, Color: c.color}
			c.orderReqC <- &orderReq
			b = true
			count = 0
		case rec := <-c.orderC:
			if b {
				startSn = rec.lsn
				b = false
			}
			count++
		}
	}
}

func (c *colorService) handleOrderResponses() error {
	for orderResp := range c.orderRespC {
		for i := int64(0); i < orderResp.NumOfRecords; i++ {
			c.Lock()
			rec, ok := c.lsnToRecord[orderResp.StartLsn+i]
			delete(c.lsnToRecord, orderResp.StartLsn+i)
			c.Unlock()
			if !ok {
				continue
			}
			err := c.disk.Assign(0, orderResp.StartLsn+i, 1, orderResp.StartGsn+i)
			if err != nil {
				return err
			}
			rec.gsnWait <- orderResp.StartGsn + i
			c.comRecC <- &spb.CommittedRecord{Color: c.color, Gsn: orderResp.StartGsn + i, Record: rec.r}

		}
	}
	return nil
}

func (c *colorService) subscribe(gsn int64, comRecC chan *pb.CommittedRecord) int64 {
	go c.readStartingFromGsn(gsn, comRecC)

	c.subCsMu.Lock()
	c.subCsCnt++
	c.subCs[c.subCsCnt] = comRecC
	ret := c.subCsCnt
	c.subCsMu.Unlock()

	return ret
}

func (c *colorService) removeSubscription(subID int64) {
	c.subCsMu.Lock()
	delete(c.subCs, subID)
	c.subCsMu.Unlock()
}

func (c *colorService) readStartingFromGsn(gsn int64, comRecC chan *pb.CommittedRecord) {
	c.secGsnMu.Lock()
	to := c.secGsn
	c.secGsnMu.Unlock()
	for i := gsn; i < to; i++ {
		r, err := c.disk.ReadGSN(1, i)
		if err != nil {
			continue
		}
		comRec := pb.CommittedRecord{Record: r, Gsn: i, Color: c.color}
		comRecC <- &comRec
	}
}
