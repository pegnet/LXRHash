// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package lxr

import "testing"

var lx LXRHash
var oprhash []byte

func init() {
	lx.Init(0xfafaececfafaecec, 25, 256, 5)
	oprhash = lx.Hash([]byte(`Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nunc dapibus pretium urna, mollis aliquet elit cursus ac. Sed sodales, erat ut volutpat viverra, ante urna pretium est, non congue augue dui sed purus. Mauris vitae mollis metus. Fusce convallis faucibus tempor. Maecenas hendrerit, urna eu lobortis venenatis, neque leo consequat enim, nec placerat tellus eros quis diam. Donec quis vestibulum eros. Maecenas id vulputate justo. Quisque nec feugiat nisi, lacinia pulvinar felis. Pellentesque habitant sed.`))
}

func BenchmarkHash(b *testing.B) {
	nonce := []byte{0, 0}
	for i := 0; i < b.N; i++ {
		nonce = nonce[:0]
		for j := i; j > 0; j = j >> 8 {
			nonce = append(nonce, byte(j))
		}
		no := append(oprhash, nonce...)
		h := lx.Hash(no)

		var difficulty uint64
		for i := uint64(0); i < 8; i++ {
			difficulty = difficulty<<8 + uint64(h[i])
		}
	}
}
