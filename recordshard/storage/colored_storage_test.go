package storage

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewColoredStorage(t *testing.T) {
	cs, err := NewColoredStorage("temp")
	require.NoError(t, err)
	require.Equal(t, cs.path, "temp")
	require.NotNil(t, cs.colorToDisk)
	defStorage, ok := cs.colorToDisk[0]
	require.True(t, ok)
	require.Equal(t, defStorage.path, "temp/0")
}

func TestColoredStorage_AddNewDisk(t *testing.T) {
	cs, err := NewColoredStorage("temp")
	require.NoError(t, err)
	disk, err := cs.AddNewDisk(12)
	require.NoError(t, err)
	require.Equal(t, disk.path, "temp/12")
	disk2, ok := cs.colorToDisk[12]
	require.True(t, ok)
	require.Same(t, disk2, disk)
}

func TestColoredStorage_GetDiskFromColor(t *testing.T) {
	cs, err := NewColoredStorage("temp")
	require.NoError(t, err)
	disk, err := cs.AddNewDisk(12)
	require.NoError(t, err)
	disk2, err := cs.getDiskFromColor(12)
	require.NoError(t, err)
	require.Same(t, disk, disk2)
}
