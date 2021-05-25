package shard

import (
	"context"
	pb "github.com/nathanieltornow/ostracon/shard/shardpb"
	"github.com/sirupsen/logrus"
	"io"
)

func (s *Shard) Append(ctx context.Context, request *pb.AppendRequest) (*pb.CommittedRecord, error) {
	s.diskMu.Lock()
	lsn, err := s.disk.Write(request.Record)
	s.diskMu.Unlock()
	if err != nil {
		return nil, err
	}

	var gsn int64
	if !s.isRoot {
		gsn = s.WaitForGsn(lsn)
	} else {
		s.lsn += 1
		gsn = s.lsn
	}

	s.diskMu.Lock()
	err = s.disk.Sync()
	if err != nil {
		return nil, err
	}
	err = s.disk.Commit(lsn, gsn)
	if err != nil {
		return nil, err
	}
	err = s.disk.Sync()
	if err != nil {
		return nil, err
	}
	s.diskMu.Unlock()

	return &pb.CommittedRecord{Gsn: gsn, Record: request.Record}, nil
}

func (s *Shard) GetOrder(stream pb.Shard_GetOrderServer) error {
	if s.isRoot && s.isSequencer {
		for {
			req, err := stream.Recv()
			logrus.Infoln(req.StartLsn, req.NumOfRecords)
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}

			s.lsnMu.Lock()
			res := pb.OrderResponse{StartGsn: s.lsn, StartLsn: req.StartLsn}
			s.lsn += req.NumOfRecords
			s.lsnMu.Unlock()

			if err := stream.Send(&res); err != nil {
				return err
			}
		}
	}

	return nil
}
