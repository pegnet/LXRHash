package lxr

const (
	HBits = 0x20
	HMask = HBits - 1
)

type LXRHash struct {
	ByteMap [Mapsiz]byte // Integer Offsets
}

// Hash()
// Takes a source of bytes, returns a 32 byte (256 bit) hash
func (w LXRHash) Hash(src []byte) []byte {

	// Keep the 32 byte intermediate result as int64 values until reduced.
	var hashes [HBits]int64
	// The intital offset into the lookup table is the length of the input.
	// This prevents concatenation attacks, which adds to the protection from
	// the reduction pass.
	var offset = int64(len(src))
	// We keep a series of previous states, and roll them along through each
	// byte of source processed.
	var last1, last2, last3 int64

	v := byte(offset)
	// Pass through the source bytes, building up lastX values, hashes[], and offset
	for i, v2 := range src {
		// Take the byte from source (v2) and map it through the lookup table
		// using the offset being maintained, and the rolling lastX values
		v = w.ByteMap[(offset^int64(v2)^last1^last2^last3)&MapMask] ^ v

		// Roll the set of last values, leaving lingering influences from past
		// values.
		last3 = last2>>2 ^ last3
		last2 = last1<<3 ^ last2
		last1 = int64(v2) ^ last1<<1

		// Set one of the hashes[] using the last rolling value, the input byte v2,
		// the mapped byte v, and the previous hashes[] value
		h := hashes[i&HMask]
		hashes[i&HMask] = last3 ^ int64(v^v2) ^ h

		// combine the l values, the previous offset, and the hashes[i]
		offset = last1<<7 ^ last2<<3 ^ last3<<9 ^ offset<<8 ^ offset>>1 ^ h
	}

	// Reduction pass
	// Done by Interating over hashes[] to produce the bytes[] hash
	//
	// At this point, we have HBits of state in hashes.  We need to reduce them down to a byte,
	// And we do so by doing a bit more bitwise math, and mapping the values through our byte map.

	var bytes [HBits]byte

	// Roll over all the hashes (32 int64 values)
	for i, h := range hashes {
		// Map each h using the offset and rolling values
		v := w.ByteMap[(offset^h^last1^last2^last3)&MapMask]
		// Roll the last values
		last3 = last2>>2 ^ last3
		last2 = last1<<3 ^ last2
		last1 = h ^ last1<<1

		// Set a byte
		bytes[i] = w.ByteMap[(int64(v)^offset)&MapMask]

		// combine the l values, the previous offset, and the hashes[i]
		offset = last1<<7 ^ last2<<3 ^ last3<<9 ^ offset<<8 ^ offset>>1 ^ h
	}

	// Return the resulting hash
	return bytes[:]
}
