package testing_test

import lxr "github.com/pegnet/LXRHash"

const (
	Seed        = uint64(0xFAFAECECFAFAECEC) // The seed defines a "hash space".
	MapSizeBits = uint64(20)                 // Default table size
	Passes      = uint64(5)                  // Default number of shuffles of the tables
	HashSize    = uint64(256)                // Default hash size.
)

var LX lxr.LXRHash
