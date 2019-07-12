// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package testing_test

import (
	"math/rand"
	"testing"
	"time"

	. "github.com/pegnet/LXRHash"
)

func TestAll(t *testing.T) {
	rand.Seed(123412341234)

	Lxrhash.Init(Seed, MapSizeBits, HashSize, Passes)

	Gradehash{}.PrintHeader()

	numTests := 1
	for i := 0; i < numTests; i++ {
		go BitChangeTest()
		go BitCountTest()
		go DifferentHashes()
		go AddByteTest()
	}

	time.Sleep(550 * time.Second)

}
