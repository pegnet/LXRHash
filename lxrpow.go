// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package lxr

// LxrPoW() returns a 64 byte value indicating the proof of work of a hash
// This is designed to ensure we use the cryptography of sha256, but require
// the random byte access limits of LXRHash
// The bigger uint64, the more PoW it represents.  The first byte is the
// number of leading bytes of FF, followed by the reset of the hash.
func (lx LXRHash) LxrPoW(hash []byte) (LHash []byte, pow uint64) {

	LHash = append(LHash, hash...)
	// If the LxrPoW isn't at least 8 bytes, there is little point in calculating a difficulty
	if len(LHash) < 8 {
		return LHash, 0
	}

	// This function uses a uint64 to index into the ByteMap.  Note MapSize is a power of 2, so MapSize-1 is a mask
	B := func(v uint64) uint64 { return uint64(lx.ByteMap[v&(lx.MapSize-1)]) }

	var state uint64
	// We assume the LxrPoW is a good cryptographic LxrPoW, like Sha256.  Initalize state with the first 8 bytes of the LxrPoW
	for i := 0; i < 8; i++ {
		state = state<<8 ^ B(state^uint64(LHash[i]))
	}

	// Make a number of passes through the LxrPoW
	for i := 0; i < 30; i++ {
		for j := len(LHash) - 1; j >= 0; j-- {
			state, LHash[j] = state<<17^state>>7^B(state+uint64(LHash[j])), byte(state)
		}
	}
	return LHash, lx.PoW(LHash)
}

// Return a uint64 as the difficulty.  Where the most significant bit is bit 0:
//   - bits 0-3 is the Count of leading bytes of "FF"
//   - bits 4-63 are the following bits of the hash
// What this does is allow difficulty to be computed by at least the first 60 bits of the hash, and as much as
// 15*8+60 bits, or 180 bits, while representing the difficulty is a simple unsigned 64 bit int.
//
// Larger values represent larger difficulties.
func (lx LXRHash) PoW(hash []byte) uint64 {
	cnt := uint64(0) // Count leading bytes of 0xFF
	idx := 0         // idx is the index of bytes to collect into the difficulty
	for hash[idx] == 0xff {
		idx++
	}
	cnt = uint64(idx) // At this point, idx is both the count of leading 0xFF bytes, and pointer to next byte
	pow := uint64(0)  // We are now ready to collect 8 bytes of hash following the leadding 0xFF bytes
	for i := 0; i < 8 && idx < len(hash); i++ {
		pow = pow << 8
		pow = pow ^ uint64(hash[idx])
		idx++
	}
	pow = pow>>4 ^ cnt<<60 // To put the 0xFF byte count in the top 4 bits, shift pow right to make room, and cnt left 60 bits and combine
	return pow
}
