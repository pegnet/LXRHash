# LXRHash
Lookup XoR hash
---------
This is a simple hash algorithm that takes advantage of a lookup table of randomized sets of bytes.  This lookup table 
consists of any number of 256 byte tables combined and sorted in one large table.  We then index into this large 
table to effectively look through the entire combination of tables as we translate the source data into a hash.

All parameters are specified.  The size of the lookup table (in numbers of 256 byte tables), the seed used to shuffle
the lookup table, the number of rounds to shuffle the table, and the size of the resulting hash.

This hash function has some interesting qualities.  Very large lookup tables will blow the cache on pretty much any 
processor or computer architecture. The number of bytes in the resulting hash can be increased for more security, 
without any more processing time.  Note, while this approach *can* be fast, this implemenation isn't.  The use case 
is aimed at Proof of Work (PoW), not cryptographic hashing.
  
The lookup 
-------
table has equal numbers of every byte value, but has them randomized over the whole table.  When hashing, the bytes from 
the source data are used to build offsets and state that are used to create the hash.

In developing this hash, the goal was to produce very randomized hashes as outputs, with a strong avalanche response to 
any change to any source byte.

LRXHash was only developed this hash as a thought experiment, and yeilds some interesting qualities.

* the lookup table can be any size, so making a version that is ASIC resistant is possible by using very big lookup tables.  Such tables blow the processor caches on CPUs and GPUs, making the speed of the hash dependent on random access of memory, not processor power.  Using 1 GB lookup table, a very fast ASIC improving hashing is limited to about ~1/3 of the computational time for the hash.  2/3 of the time is spent waiting for memory access.  
* at smaller lookup table sizes where processor caches work, LXRHash can be modified to be very fast.
* LXRHash would be an easy ASIC design as it only uses counters, decrements, XORs, and shifts. 
* the hash is trivially altered by changing the size of the lookup table, the seed, size of the hash produced. Change any parameter and you change the space from which hashes are produced.

While this hash may be reasonable for use as PoW in mining on an immutable ledger that provides its own security, 
not nearly enough testing has been done to use as a fundamental part in cryptography or security.  For fun, it 
would be cool to do such testing.

The actual implementation is very small, and see the code for the most accurate source. The code is presented here without comments to illustrate its small size (Hash() is ~ 29 lines of go) :
```go
type LXRHash struct {
	ByteMap     []byte // Integer Offsets
	MapSize     uint64 // Size of the translation table
	MapSizeBits uint64 // Size of the ByteMap in Bits
	Passes      uint64 // Passes to generate the rand table
	Seed        uint64 // An arbitrary number used to create the tables.
	HashSize    uint64 // Number of bytes in the hash
}
func (lx LXRHash) Hash(src []byte) []byte {
	hs := make([]uint64, lx.HashSize)
	var as = lx.Seed
	var s, s2 [7]uint64
	mk := lx.MapSize - 1
	step := func(i uint64, v2 uint64) {
		s[0] = s[0] ^ as ^ v2 ^ uint64(lx.ByteMap[(as^v2<<9)&mk])<<4
		for i := len(s) - 1; i >= 0; i-- {
			if i > 0 {
				s[i] = s[i-1]<<23 ^ s[i-1]>>5 ^ s[i]<<17 ^ s[i]>>3 ^ uint64(lx.ByteMap[(s[i]^v2<<9)&mk])<<11
			}
			as = s[i]<<29 ^ s[i]>>5 ^ as<<17 ^ as>>1 ^ uint64(lx.ByteMap[(s[3]^v2<<9)&mk])<<13
			s[0], s[1], s[2], s[3], s[4], s[5], s[6] = s[4], s[5], s[2], s[6], s[3], s[0], s[1]
		}
		s, s2 = s2, s
	}
	for i, v2 := range src {
		idx := uint64(i)
		step(idx, uint64(v2))
		hash := hs[idx%lx.HashSize]
		hs[idx%lx.HashSize] = as ^ hash<<21 ^ hash>>1
	}
	bytes := make([]byte, lx.HashSize)
	for i, h := range hs {
		step(uint64(i), h)
		idx2 := (s[0] ^ as) & mk
		bytes[i] = lx.ByteMap[idx2]
	}
	return bytes
}
```

The current code has added 256 bytes of stages to accumulate more state quickly as the code moves through the source.  