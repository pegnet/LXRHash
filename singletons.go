package lxr

import (
	"fmt"
	"sync"
)

// The goal of instances is to provide a way for multiple packages to use LXR without
// instantiating multiple bytemaps in memory or having to share references

var instanceMtx sync.Mutex
var instances map[string]*LXRHash
var counter map[string]uint64

func init() {
	instances = make(map[string]*LXRHash)
	counter = make(map[string]uint64)
}

// Init provides access to shared instances of LXRHash without having to instantiate multiple bytemaps.
// Two separate calls to Init() will result in a reference to the same object.
func Init(seed, bitsize, hashsize, passes uint64) *LXRHash {
	if bitsize < 8 {
		panic("bitsize must be at least 8")
	}

	instanceMtx.Lock()
	defer instanceMtx.Unlock()

	id := fmt.Sprintf("%d-%d-%d-%d", seed, bitsize, hashsize, passes)

	counter[id]++

	if instance, ok := instances[id]; ok {
		return instance
	}

	lxr := new(LXRHash)
	lxr.Verbose(true)
	lxr.Init(seed, bitsize, hashsize, passes)
	instances[id] = lxr
	return lxr
}

// Release releases a singleton. If all references to the singleton have been released, the singleton is destroyed
// and can be garbage collected
func Release(hash *LXRHash) {
	if hash == nil {
		return
	}
	instanceMtx.Lock()
	defer instanceMtx.Unlock()

	id := fmt.Sprintf("%d-%d-%d-%d", hash.Seed, hash.MapSizeBits, hash.HashSize*8, hash.Passes)
	test, exists := instances[id]
	if !exists || test != hash {
		panic("tried to release a non-singleton instance")
	}
	counter[id]--
	if counter[id] == 0 {
		delete(counter, id)
		delete(instances, id)
	}
}
