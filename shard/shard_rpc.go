package shard

import (
	"context"
	pb "github.com/nathanieltornow/ostracon/shard/shardpb"
	"github.com/sirupsen/logrus"
	"io"
)

func (s *Shard) Append(ctx context.Context, request *pb.AppendRequest) (*pb.CommittedRecord, error) {
	lsn, err := s.disk.Write(request.Record)

	if err != nil {
		return nil, err
	}
	s.snMu.Lock()
	s.sn += 1
	s.snMu.Unlock()
	var gsn int64
	if !s.isRoot {
		gsn = s.WaitForGsn(lsn)
	} else {
		gsn = s.sn
	}
	err = s.disk.Sync()
	if err != nil {
		logrus.Infoln("1")

		return nil, err
	}
	err = s.disk.Commit(lsn, gsn)
	if err != nil {
		logrus.Infoln("2")

		return nil, err
	}
	err = s.disk.Sync()
	if err != nil {
		logrus.Infoln("3")
		return nil, err
	}
	return &pb.CommittedRecord{Gsn: gsn, Record: request.Record}, nil
}

func (s *Shard) GetOrder(stream pb.Shard_GetOrderServer) error {
	if s.isRoot && s.isSequencer {
		for {
			req, err := stream.Recv()
			logrus.Infoln("Received OrderReq", req.StartLsn, req.NumOfRecords)
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}

			s.snMu.Lock()
			res := pb.OrderResponse{StartGsn: s.sn, StartLsn: req.StartLsn, NumOfRecords: req.NumOfRecords}
			s.sn += req.NumOfRecords
			s.snMu.Unlock()

			if err := stream.Send(&res); err != nil {
				return err
			}
		}
	}

	return nil
}
