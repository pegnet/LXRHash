// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package lxr

import (
	"bytes"
	"errors"
	"fmt"
)

// LXRHash holds one instance of a hash function with a specific seed and map size
type LXRHash2 struct {
	ByteMap     []byte // Integer Offsets
	MapSize     uint64 // Size of the translation table
	MapSizeBits uint64 // Size of the ByteMap in Bits
	Passes      uint64 // Passes to generate the rand table
	Seed        uint64 // An arbitrary number used to create the tables.
	HashSize    uint64 // Number of bytes in the hash
	FirstIdx    uint64 // First Index used by LXRHash. (variance measures distribution of ByteMap access)
	verbose     bool
}

// Returns just the 32 byte hash.
func (lx *LXRHash2) Hash(src []byte) []byte {
	h, _ := lx.HashValidate(src, nil)
	return h[:32]
}

// Hash takes the arbitrary input and returns what amounts to a 256 byte hash.  32 bytes are the literal hash
// of the data using the ByteMap, and the remaining 196 bytes are the inputs from the ByteMap required to recompute
// the hash.
//
// Takes the source document, and the 256 byte hash.  If the hash is nil, the hash is computed.  If the hash is given
// then we return the hash if it validates, or a nil if the hash fails validation.
func (lx *LXRHash2) HashValidate(src []byte, hash []byte) ([]byte, error) {
	// The hash to validate must be nil (so we compute it) or length 256 if we are validating the hash.
	if hash != nil && len(hash) != 256 {
		return nil, errors.New("hash is the incorrect length")
	}
	// Since MapSize is specified in bits, the index mask is the size-1
	mk := lx.MapSize - 1
	// If we are building a hash, then we will be collecting the bytes into a verfication list.
	var vlist []byte

	// Pick a function; assume that we are building a hash, not validating a hash
	B := func(v uint64) byte { b := lx.ByteMap[v&mk]; vlist = append(vlist, b); return b }
	if hash != nil {
		h := hash[32:] // h will point to the validation bytes of the hash
		// But if it turns out we are validating a hash, not building one, then use the validation byte source
		B = func(v uint64) (b byte) { b, h = h[0], h[1:]; return b }
	}

	// The offset variable is all the state we really need for PoW.  The more complex state used in
	// the origonal LXRHash makes it a better cryptographic hash, but allows computation to play a bigger role
	// in PoW.
	//
	// That said, we have to initialize the offset variable with the full state of the data from the source.  We
	// are going to make one pass through the data and reduce it to 32 bytes.  Then make a pass through the result
	// (data the 32 bytes, and the 8 byte offset) to ensure even the last bit has an impact on all bits.
	var offset uint64
	data := make([]byte, int(lx.HashSize))
	s := 0
	d := 0
	for i := 0; i < len(src); i++ {
		s = s % len(src)
		d = d % int(lx.HashSize)
		offset = offset<<11 ^ offset>>1 ^ uint64(src[s])<<40 ^ uint64(src[s])<<16
		data[d] = byte(offset) ^ byte(offset>>8) ^ byte(offset>>16) ^ byte(offset>>24) ^ data[d]
		s++
		d++
	}
	// Because we are now going to process just the 32 bytes that step one reduced our source down, we get a
	// standard sized byte stream for the validation, no matter how much data we are hashing.
	for i, b := range data {
		offset = offset<<11 ^ offset>>1 ^ uint64(b)<<32 ^ uint64(b)
		data[i] = byte(offset)
	}

	// Apply the validation stream to the hashing mix.  Also ensure the validation stream is well mixed into
	// the offset state.
	p := 0
	dv := uint64(0)
	for i := 0; i < 7*int(lx.HashSize); i++ {
		p = p % len(data)
		d := data[p]
		b := uint64(B(offset + uint64(d)))
		dv = dv<<7 ^ dv>>1 ^ b
		d = d ^ byte(dv) ^ byte(offset) ^ byte(offset>>8) ^ byte(offset>>16) ^ byte(offset>>24) //^ byte(offset^offset>>11^offset>>33)
		offset = offset<<7 ^ offset>>1 ^ dv<<3 ^ dv<<23 ^ uint64(d)
		data[p] = d
		p++
	}

	// Do one more pass to ensure that a change to the end of the validation stream has an impact on the entire hash
	// result, and not just on a byte and its following result.
	for i, b := range data {
		offset = offset<<11 ^ offset>>1 ^ uint64(b)<<32 ^ uint64(b)
		data[i] = byte(offset)
	}

	// If we are validating the hash, then return either a nil or the hash we were given.
	if hash != nil {
		if !bytes.Equal(data[:], hash[:32]) { // The hash didn't validate, so return a nil
			return append(data[:], hash[32:]...), fmt.Errorf("expected %x but got %x", hash[:32], data[:])
		}
		// The hash did validate, return it as is.
		return hash, nil
	}

	// If building a list, return the 32 byte hash, with the vlist appended to the hash.
	return append(data[:], vlist...), nil
}
