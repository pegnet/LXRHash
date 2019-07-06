package testing_test

import (
	. "github.com/pegnet/LXRHash"
	"math/rand"
	"testing"
	"time"
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

	time.Sleep(180 * time.Second)

}
