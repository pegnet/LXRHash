package testing_test

import (
	"fmt"
	. "github.com/pegnet/LXRHash"
	"testing"
)

// Create all the tables for a range of parameters for the LXRHash
// These tables are used to test and characterize the LXRHash

// TestGenerateTables()
// Generate all the map files for a particular seed, and number of passes
func TestGenerateTables(t *testing.T) {

	// Create all size tables with a given seed, up to 10 passes and for sizes upto 4 GB
	for i := uint64(8); i < 33; i++ {
		for passes := uint64(5); passes <= 7; passes++ {
			meg := float64(uint64(1)<<i) / 1000000
			fmt.Printf("Processing Map of %12.4f MB, %d bits\n", meg, i)
			lxrHash := LXRHash{}
			lxrHash.Init(Seed, i, HashSize, passes)
		}
	}
}
