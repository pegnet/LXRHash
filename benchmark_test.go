package lxr

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"sync"
	"testing"
	"time"
)

func Test_benchMiner(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	dupe := make(map[string]bool)

	var count uint64
	var realCount uint64
	base := []byte{0xab, 0xab, 0xab}
	go benchMiner(ctx, 0x55, &count, base, func(in []byte) []byte {
		str := hex.EncodeToString(in)
		if dupe[str] {
			t.Errorf("duplicate nonce: %s", str)
		}
		dupe[str] = true

		fmt.Printf("%x\n", in)
		if !bytes.Equal(base, in[:len(base)]) {
			t.Errorf("supplied base invalid. want = %x, got = %x", base, in[:len(base)])
		}
		realCount++
		return in
	})

	time.Sleep(time.Millisecond)
	cancel()

	if realCount != count {
		t.Errorf("count mismatch. realCount = %d, count = %d", realCount, count)
	}

	time.Sleep(time.Millisecond)
	snapshot := count
	time.Sleep(time.Millisecond)

	if snapshot != count {
		t.Errorf("goroutine keeps running. snapshot = %d, count = %d", snapshot, count)
	}
}

func Test_benchMiner_concurrent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var mtx sync.Mutex
	dupe := make(map[string]bool)

	var count uint64
	base := []byte{0xab, 0xab, 0xab}
	for i := 0; i < 4; i++ {
		go benchMiner(ctx, byte(i), &count, base, func(in []byte) []byte {
			mtx.Lock()
			str := hex.EncodeToString(in)
			if dupe[str] {
				t.Errorf("duplicate nonce: %s", str)
			}
			dupe[str] = true
			mtx.Unlock()

			return nil
		})
	}

	time.Sleep(time.Millisecond * 500)
	cancel()
}

func Test_benchFunc(t *testing.T) {
	testFunc := func(_ []byte) []byte {
		return nil
	}

	// test duration
	start := time.Now()
	hashes, duration := benchFunc(context.Background(), time.Millisecond*100, 1, testFunc)
	realDuration := time.Since(start)

	if duration > realDuration {
		t.Errorf("duration reporting wrong. got = %s, real = %s", duration, realDuration)
	} else if duration == 0 {
		t.Errorf("zero duration return")
	}

	if hashes == 0 {
		t.Errorf("no hashes calculated")
	}

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)

	start = time.Now()
	go func() {
		_, duration = benchFunc(ctx, time.Second*2, 1, testFunc)
		wg.Done()
	}()

	time.Sleep(time.Millisecond * 500)
	cancel()
	cancelDuration := time.Since(start)
	wg.Wait()
	realDuration = time.Since(start)

	if realDuration-cancelDuration > time.Second {
		t.Errorf("Cancelling took too long. cancelDuration = %s, realDuration = %s", cancelDuration, realDuration)
	}

	if duration-cancelDuration > time.Second {
		t.Errorf("Cancelling took too long. cancelDuration = %s, benchDuration = %s", cancelDuration, duration)
	}
}
