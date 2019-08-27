package lxr

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func Benchmark_GenerateTable(b *testing.B) {
	b.Run("slow", func(b *testing.B) {
		ByteMap := make([]byte, b.N)
		period := time.Now().Unix()
		for i := range ByteMap {
			if (i+1)%1000 == 0 && time.Now().Unix()-period > 10 {
				println(" Index ", i+1, " of ", len(lx.ByteMap))
				period = time.Now().Unix()
			}
			ByteMap[i] = byte(i)
		}
	})
	b.Run("fast", func(b *testing.B) {
		ByteMap := make([]byte, b.N)
		for i := range ByteMap {
			ByteMap[i] = byte(i)
		}
	})
}

func (lx *LXRHash) OldWriteTable(filename string) {
	os.Remove(filename)

	// open output file
	fo, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	// write a chunk
	if _, err := fo.Write(lx.ByteMap[:]); err != nil {
		panic(err)
	}

}

func compareWrite(t *testing.T, MapSizeBits uint64) {
	l := new(LXRHash)

	MapSize := uint64(1) << MapSizeBits
	l.ByteMap = make([]byte, int(MapSize))

	l.HashSize = (HashSize + 7) / 8
	l.MapSize = MapSize
	l.MapSizeBits = MapSizeBits
	l.Seed = Seed
	l.Passes = Passes
	l.GenerateTable()

	o, err := ioutil.TempFile("", "bytemapold")
	defer os.Remove(o.Name())
	if err != nil {
		panic(err)
	}
	o.Close()

	l.OldWriteTable(o.Name())

	n, err := ioutil.TempFile("", "bytemapnew")
	defer os.Remove(n.Name())
	if err != nil {
		panic(err)
	}
	n.Close()

	l.WriteTable(n.Name())

	a, err := ioutil.ReadFile(o.Name())
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadFile(n.Name())
	if err != nil {
		panic(err)
	}

	if !bytes.Equal(a, b) {
		t.Errorf("mismatch for %d bits. old = %32x, new = %32x", MapSizeBits, a, b)
	}
}

func TestLXRHash_WriteTable(t *testing.T) {
	compareWrite(t, 8)
	compareWrite(t, 12)
	compareWrite(t, 16)
	compareWrite(t, 20)
}
