// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package lxr

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

func (lx *LXRHash2) Hash(src []byte) []byte {
	h, _ := lx.HashValidate(src, nil, false)
	return h
}

// Hash takes the arbitrary input and returns the resulting hash of length HashSize
func (lx *LXRHash2) HashValidate(src []byte, vlist []byte, validate bool) (rt []byte, list []byte) {
	// Since MapSize is specified in bits, the index mask is the size-1
	mk := lx.MapSize - 1

	B := func(v uint64) byte { b := lx.ByteMap[v&mk]; list = append(list, b); return b }
	if validate {
		B = func(v uint64) (b byte) { b, vlist = vlist[0], vlist[1:]; return b }
	}
	var offset uint64
	data := make([]byte, int(lx.HashSize))
	s := 0
	d := 0
	for i := 0; i < len(src)*3; i++ {
		if s >= len(src) {
			s = 0
		}
		if d >= int(lx.HashSize) {
			d = 0
		}
		offset = offset<<11 ^ offset>>1 ^ uint64(src[s])<<33 ^ uint64(src[s])>>7
		data[d] = byte(offset) ^ data[d]
		s++
		d++
	}
	p := 0
	for i := 0; i < 8*int(lx.HashSize); i++ {
		data[p] = data[p] ^ B(offset+uint64(data[p])) ^ byte(offset^offset>>11^offset>>33)
		offset = offset<<15 ^ offset>>7 ^ uint64(data[p])<<3 ^ uint64(data[p])<<23
		p++
		if p >= int(lx.HashSize) {
			p = 0
		}
	}

	return data[:], list
}
