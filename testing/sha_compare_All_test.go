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

	numTests := 2
	for i := 0; i < numTests; i++ {
		go BitChangeTest()
		go BitCountTest()
		go DifferentHashes()
		go AddByteTest()
	}

	time.Sleep(5500 * time.Second)

}
