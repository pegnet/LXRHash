package lxr

// The goal of instances is to provide a way for multiple packages to use LXR without
// instantiating multiple bytemaps in memory or having to share references

var instances map[uint64]*LXRHash

func init() {
	instances = make(map[uint64]*LXRHash)
}

// Init provides access to shared instances of LXRHash without having to instantiate multiple bytemaps.
// Two separate calls to Init(n) will result in a reference to the same object.
// LXRHash will be instantiated with the package defaults of:
// 	* Seed: 0xFAFAECECFAFAECEC
// 	* Hash Size: 256
// 	* Passes: 5
func Init(bitsize uint64) *LXRHash {
	if bitsize < 8 {
		panic("bitsize must be at least 8")
	}

	if instance, ok := instances[bitsize]; ok {
		return instance
	}

	lxr := new(LXRHash)
	lxr.Verbose(true)
	lxr.Init(Seed, bitsize, HashSize, Passes)
	instances[bitsize] = lxr
	return lxr
}

// Release dereferences the shared instance inside this package, allowing the garbage collector to free the memory.
// Please note that any existing references to the instance outside this package will keep it alive
func Release(bitsize uint64) {
	delete(instances, bitsize)
}
