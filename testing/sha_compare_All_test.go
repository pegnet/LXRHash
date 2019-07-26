// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package testing_test

import (
	"testing"
	"time"
)

func TestAll(t *testing.T) {
	LX.Init(Seed, MapSizeBits, HashSize, Passes)

	Gradehash{}.PrintHeader()

	numTests := 1
	for i := 0; i < numTests; i++ {
		go BitChangeTest()
		go BitCountTest()
		go DifferentHashes()
		go AddByteTest()
	}

	time.Sleep(20 * time.Second)

}
