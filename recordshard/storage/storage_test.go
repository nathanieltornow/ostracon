package storage

import (
	"os"
	"testing"
)

func TestNewStorage(t *testing.T) {
	record := "test"
	p, err := NewStorage("tmp", 0, 2, 1000)
	check(t, err)
	if p == nil {
		t.Errorf("Get nil segment on creating")
	}
	l, err := p.Write(record)
	check(t, err)
	if l != 0 {
		t.Errorf("Write error: expect ssn %v, get %v", len(record), l)
	}
	r, err := p.ReadLSN(0, 0)
	check(t, err)
	if r != record {
		t.Errorf("Read error: expect '%v', get '%v'", record, r)
	}
	err = p.Assign(0, 0, 1, 100)
	check(t, err)

	r, err = p.ReadGSN(100)
	check(t, err)
	if r != record {
		t.Errorf("Read error: expect '%v', get '%v'", record, r)
	}

	err = os.RemoveAll("tmp")
	check(t, err)
}

func BenchmarkName(b *testing.B) {
	p, _ := NewStorage("tmp", 0, 2, 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.Write("Hallo")
		_, _ = p.Write("Hallo")
		_, _ = p.Write("Hallo")
		_, _ = p.Write("Hallo")

		_ = p.Assign(0, 0, 4, int64(i))
		_, _ = p.ReadGSN(int64(i))
	}

}
