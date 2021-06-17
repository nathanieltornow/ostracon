package storage

import (
	"fmt"
	"sync"
)

type ColoredStorage struct {
	sync.Mutex
	colorToDisk map[int64]*Storage
	path        string
}

func NewColoredStorage(path string) *ColoredStorage {
	cs := new(ColoredStorage)
	cs.colorToDisk = make(map[int64]*Storage)
	return cs
}

func (cs *ColoredStorage) AddNewDisk(color int64) (*Storage, error) {
	cs.Lock()
	_, ok := cs.colorToDisk[color]
	cs.Unlock()
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

func (cs *ColoredStorage) GetDiskFromColor(color int64) (*Storage, error) {
	cs.Lock()
	disk, ok := cs.colorToDisk[color]
	cs.Unlock()
	if ok {
		return disk, nil
	}
	return cs.AddNewDisk(color)
}
