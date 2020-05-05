// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package lxr

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"testing"
	"time"
)

var Found chan solution

type solution struct {
	time    time.Time
	total   uint64
	cnt     int
	comment string
	hash    []byte
	lhash   []byte
	diff    uint64
}

func TstHash(comment string, size int, lxrPoW bool) {
	powCnt := 0
	var pow, lastPow uint64

	var data string
	for i := 0; i < size; i++ {
		data += "x"
	}

	v := sha256.Sum256([]byte(data))

	var lhash []byte
	start := time.Now()
	for i := uint64(0); ; i++ {
		h1 := sha256.Sum256(append(v[:], []byte(strconv.Itoa(int(i)))...))
		hash := h1[:]
		if lxrPoW {
			lhash, pow = lx.LxrPoW(hash[:])
		} else {
			pow = lx.PoW(hash[:])
			lhash = hash
		}
		if pow > lastPow {
			lastPow = pow
			powCnt++
			Found <- solution{time.Now(), i, powCnt, comment, hash, lhash, pow}
		}
		if i&0xFF == 0 {
			if time.Now().Unix()-start.Unix() > 5 {
				start = time.Now()
				Found <- solution{comment: comment, time: start, total: i, cnt: 0, hash: hash, lhash: lhash, diff: pow}
			}
		}
	}
}

func TestPoW(t *testing.T) {
	start := time.Now()
	Found = make(chan solution, 1000)
	go TstHash("lxr 10  ", 10, true)
	go TstHash("hsh 10  ", 10, false)
	go TstHash("lxr 1000", 1000, true)
	go TstHash("hsh 1000", 1000, false)
	for {
		solution := <-Found
		seconds := uint64(solution.time.Sub(start).Seconds())
		if seconds == 0 {
			seconds = seconds + 1
		}

		if solution.cnt != 0 {
			fmt.Printf("Comment %32s Rate %10d/s cnt %4d POW %016x Hash %x, LxrPoW %x\n",
				solution.comment, solution.total/seconds,
				solution.cnt, solution.diff, solution.hash, solution.lhash)
		} else {
			fmt.Printf("Comment %32s Rate %10d/s \n",
				solution.comment, solution.total/seconds)
		}
	}
}
