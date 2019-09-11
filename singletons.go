package lxr

import "sync"

// The goal of instances is to provide a way for multiple packages to use LXR without
// instantiating multiple bytemaps in memory or having to share references

var instanceMtx sync.Mutex
var instances map[string]*LXRHash

func init() {
	instances = make(map[string]*LXRHash)
}

// Init provides access to shared instances of LXRHash without having to instantiate multiple bytemaps.
// Two separate calls to Init(n) will result in a reference to the same object.
// LXRHash will be instantiated with the package defaults of:
// 	* Seed: 0xFAFAECECFAFAECEC
// 	* Hash Size: 256
// 	* Passes: 5
func Init(Seed, bitsize, HashSize, Passes uint64) *LXRHash {
	if bitsize < 8 {
		panic("bitsize must be at least 8")
	}

	lxr := new(LXRHash)
	lxr.SetParms(bitsize,HashSize,Seed,Passes)

	key := lxr.GetFilename()

	instanceMtx.Lock()
	defer instanceMtx.Unlock()

	if instance, ok := instances[key]; ok {
		return instance
	}

	lxr.Verbose(true)
	lxr.Init(Seed, bitsize, HashSize, Passes)
	instances[key] = lxr
	return lxr
}
