# LXR256
Lookup XoR hash

This is a simple 256 bit hash that takes advantage of a lookup table of randomized sets of bytes (0-255).  The lookup table has equal numbers of every byte value, but has them randomized over the whole table.  When hashing, the bytes from the source data are used to build offsets and state that are used to create the hash.

In developing this hash, the goal was to produce very randomized hashes as outputs, with a strong avalanche response to any change to any source byte.

I only developed this hash as a thought experiment, but which none the less has some interesting qualities.

* the lookup table can be any size, so making a version that is ASIC resistant is possible
* at the current lookup table size, LXR256 is slightly faster than Sha256, at least in my tests
* at the current lookup table size, LXR256 would be very trivial to impliment as an ASIC, and would be very fast

The actual implementation is very small.  Assuming the lookup table is fixed (w.maps), the implementation follows, but look to the source code for comments and commentary on the implementation:
```go
const(
	Mapsiz    = 0x800
	MapMask   = Mapsiz - 1
	HBits     = 0x20
	HMask     = HBits - 1
)
func (w PegHash) Hash(src []byte) []byte {
	var hashes [HBits]int64
	var offset = int64(len(src))
	var last1, last2, last3 int64
	for i, v2 := range src {
		v := w.maps[(offset^int64(v2)^last1^last2^last3)&MapMask]
		last3 = last2>>2 ^ last3
		last2 = last1<<3 ^ last2
		last1 = int64(v2) ^ last1<<1
		h := hashes[i&HMask]
		hashes[i&HMask] = last3 ^ int64(v^v2) ^ h
		offset = last1<<7 ^ last2<<3 ^ last3<<9 ^ offset<<8 ^ offset>>1 ^ h
	}
	var bytes [HBits]byte
	for i, h := range hashes {
		v := w.maps[(offset^h^last1^last2^last3)&MapMask]
		last3 = last2>>2 ^ last3
		last2 = last1<<3 ^ last2
		last1 = h ^ last1<<1
		bytes[i] = w.maps[(int64(v)^offset)&MapMask]
		offset = last1<<7 ^ last2<<3 ^ last3<<9 ^ offset<<8 ^ offset>>1 ^ h
	}
	return bytes[:]
}


```
