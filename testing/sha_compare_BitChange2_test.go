// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package testing_test

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestBitChange2(t *testing.T) {
	LX.Init(Seed, MapSizeBits, HashSize, Passes)

	Gradehash{}.PrintHeader()

	numTests := 1
	for i := 0; i < numTests; i++ {
		go BitChangeTest2()
	}

	time.Sleep(20000 * time.Second)

}

var once sync.Once

func BitChangeTest2() {
	var g1 Gradehash
	var g2 Gradehash

	cnt := int64(0)

	buf := []byte("n3jg498cm89f36y3bbeu5euenekl,tlrwuw4qw4bubu6ebeer76ier6kmrkm6rrkkrem6en5beuwewew6nv9")
	bit := byte(1)
	last := time.Now().Unix()
	for x := int64(0); x < 100000000000; x++ {
		buf[int(x*7)%len(buf)] = buf[int(x*7)%len(buf)] ^ bit
		if x%13 == 0 {
			bit++
			if bit == 0 {
				bit = 1
			}
		}
		if x%100 == 0 {
			buf[int(x)%len(buf)]++
		}

		g1.Start()
		sv := sha256.Sum256(buf)
		g1.Stop()
		g1.AddHash(buf, sv[:])

		g2.Start()
		wv := LX.HashValidate(buf, nil)
		wv2 := LX.HashValidate(buf, wv)
		if wv2 == nil {
			panic("Validate Fails")
		}
		g2.Stop()
		g2.AddHash(buf, wv[:32])
		once.Do(func() {
			fmt.Println("vlist len ", len(wv))
		})
		cnt++
		if true {
			if cnt > 1000 && time.Now().Unix()-last > 4 {
				last = time.Now().Unix()
				cnt = 0

				c, r1 := g1.Report("bit-sha")
				_, r2 := g2.Report("bit-lxr")
				// Print on one line, so if we run multiple tests at the same time, we don't
				// split the output, because go will ensure one print goes out uninterrupted.
				fmt.Printf("%10s %s\n%10s %s\n\n", c, r1, " ", r2)
			}
		} else {
			if cnt > 100000 {
				cnt = 0

				c, r1 := g1.Report("bit-sha")
				_, r2 := g2.Report("bit-lxr")
				// Print on one line, so if we run multiple tests at the same time, we don't
				// split the output, because go will ensure one print goes out uninterrupted.
				fmt.Printf("%10s %s\n%10s %s\n\n", c, r1, " ", r2)
			}

		}

	}
}
