// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package testing_test

import (
	"fmt"
	"testing"

	lxr "github.com/pegnet/LXRHash"
)

// Create all the tables for a range of parameters for the LXRHash
// These tables are used to test and characterize the LXRHash

// TestGenerateTables()
// Generate all the map files for a particular seed, and number of passes
func TestGenerateTables(t *testing.T) {

	LX.Init(Seed, MapSizeBits, HashSize, Passes)

	// Create all size tables with the default seed, HashSize, and 5 passes up
	// to a table size of 20 bits
	for i := uint64(8); i < MapSizeBits; i++ {
		meg := float64(uint64(1)<<i) / 1000000
		fmt.Printf("Processing Map of %12.4f MB, %d bits\n", meg, i)
		lxrHash := lxr.LXRHash{}
		lxrHash.Init(Seed, i, HashSize, 5)
	}
}
