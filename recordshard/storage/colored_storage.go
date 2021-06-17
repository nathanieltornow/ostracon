package storage

import (
	"fmt"
	"sync"
)

type ColoredStorage struct {
	sync.RWMutex
	colorToDisk map[int64]*Storage
	path        string
}

func NewColoredStorage(path string) (*ColoredStorage, error) {
	cs := new(ColoredStorage)
	cs.colorToDisk = make(map[int64]*Storage)
	cs.path = path
	_, err := cs.AddNewDisk(0)
	return cs, err
}

func (cs *ColoredStorage) AddNewDisk(color int64) (*Storage, error) {
	cs.RLock()
	_, ok := cs.colorToDisk[color]
	cs.RUnlock()
	if ok {
		return nil, fmt.Errorf("disk for color %v is already existing", color)
	}
	disk, err := NewStorage(fmt.Sprintf("%v/%v", cs.path, color), 0, 2, 10000000)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %v", err)
	}
	cs.Lock()
	cs.colorToDisk[color] = disk
	cs.Unlock()
	return disk, nil
}

func (cs *ColoredStorage) GetCurrentLsn(color int64, primary bool) (int64, error) {
	disk, err := cs.getDiskFromColor(color)
	if err != nil {
		return 0, err
	}
	part := int32(1)
	if primary {
		part = 0
	}
	return disk.GetNextLsn(part), nil
}

func (cs *ColoredStorage) Read(gsn, color int64) (string, error) {
	disk, err := cs.getDiskFromColor(color)
	if err != nil {
		return "", err
	}
	rec, err := disk.ReadGSN(1, gsn)
	return rec, err
}

func (cs *ColoredStorage) Write(primary bool, record string, color int64) (int64, error) {
	disk, err := cs.getDiskFromColor(color)
	if err != nil {
		return 0, err
	}
	part := int32(1)
	if primary {
		part = 0
	}
	return disk.WriteToPartition(part, record)
}

func (cs *ColoredStorage) Assign(primary bool, lsn, gsn, color, length int64) error {
	disk, err := cs.getDiskFromColor(color)
	if err != nil {
		return err
	}
	part := int32(1)
	if primary {
		part = 0
	}
	return disk.Assign(part, lsn, int32(length), gsn)
}

func (cs *ColoredStorage) getDiskFromColor(color int64) (*Storage, error) {
	cs.RLock()
	disk, ok := cs.colorToDisk[color]
	cs.RUnlock()
	if ok {
		return disk, nil
	}
	return nil, fmt.Errorf("no disk for color %v", color)
}
