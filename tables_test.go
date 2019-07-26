package lxr

import (
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
