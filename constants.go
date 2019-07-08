package lxr

// Default Seed used by the PegNet, the first application to use
// LXRHash for Proof of Work
var Seed = uint64(0xFAFAECECFAFAECEC)

// Default table size used by the PegNet
var MapSizeBits = uint64(25)

// Default number of shuffles of the tables
var Passes = uint64(10)

// Default hash size.
var HashSize = uint64(256)

var Lxrhash LXRHash

func init() {
	Lxrhash.Init(Seed, MapSizeBits, HashSize, Passes)
}
