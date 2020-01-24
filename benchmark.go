package lxr

import (
	"context"
	"encoding/binary"
	"runtime"
	"sync/atomic"
	"time"
)

// BenchmarkHash will run a benchmark for the specified duration using the regular Hash function.
// Returns the number of hashes calculated and the real duration of the benchmark.
// If no goroutines are specified it will use the total number of available cores.
func (lx LXRHash) BenchmarkHash(ctx context.Context, duration time.Duration, goroutines uint) (uint64, time.Duration) {
	return benchFunc(ctx, duration, goroutines, lx.Hash)
}

// BenchmarkHash will run a benchmark for the specified duration using the FlatHash function.
// Returns the number of hashes calculated and the real duration of the benchmark.
// If no goroutines are specified it will use the total number of available cores.
func (lx LXRHash) BenchmarkFlatHash(ctx context.Context, duration time.Duration, goroutines uint) (uint64, time.Duration) {
	return benchFunc(ctx, duration, goroutines, lx.FlatHash)
}

// benchmark a specific function. cancels early if context is cancelled, otherwise runs for duration
func benchFunc(ctx context.Context, duration time.Duration, goroutines uint, f func([]byte) []byte) (uint64, time.Duration) {
	if goroutines == 0 {
		goroutines = uint(runtime.NumCPU())
	}

	if ctx == nil {
		ctx = context.Background()
	}

	myctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	var hashes uint64
	base := make([]byte, 32) // null base is ok

	start := time.Now()
	for i := 0; i < int(goroutines); i++ {
		go benchMiner(myctx, byte(i), &hashes, base, f)
	}

	<-myctx.Done()
	return hashes, time.Since(start)
}

// individual mining thread
func benchMiner(ctx context.Context, id byte, count *uint64, base []byte, f func([]byte) []byte) {
	nonce := append(base, []byte{id, 0, 0, 0, 0}...)
	pos := len(nonce) - 4
	i := uint32(0)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			binary.BigEndian.PutUint32(nonce[pos:], i)
			f(nonce)
			atomic.AddUint64(count, 1)
			i++
		}
	}
}
