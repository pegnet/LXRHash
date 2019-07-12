package lxr

// Default Seed
const (
	Seed        = uint64(0xFAFAECECFAFAECEC) // The seed defines a "hash space".
	MapSizeBits = uint64(30)                 // Default table size
	Passes      = uint64(5)                  // Default number of shuffles of the tables
	HashSize    = uint64(256)                // Default hash size.
)

var Lxrhash LXRHash

func init() {
	Lxrhash.Init(Seed, MapSizeBits, HashSize, Passes)
}
