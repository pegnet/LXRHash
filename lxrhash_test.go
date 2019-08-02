// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package lxr

import (
	"bytes"
	"encoding/hex"
	"testing"
)

var lx LXRHash
var oprhash []byte

func init() {
	lx.Init(0xfafaececfafaecec, 30, 256, 5)
	oprhash = lx.Hash([]byte(`Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nunc dapibus pretium urna, mollis aliquet elit cursus ac. Sed sodales, erat ut volutpat viverra, ante urna pretium est, non congue augue dui sed purus. Mauris vitae mollis metus. Fusce convallis faucibus tempor. Maecenas hendrerit, urna eu lobortis venenatis, neque leo consequat enim, nec placerat tellus eros quis diam. Donec quis vestibulum eros. Maecenas id vulputate justo. Quisque nec feugiat nisi, lacinia pulvinar felis. Pellentesque habitant sed.`))
}

func BenchmarkHash(b *testing.B) {
	b.Run("hash", func(b *testing.B) {
		nonce := []byte{0, 0}
		for i := 0; i < b.N; i++ {
			nonce = nonce[:0]
			for j := i; j > 0; j = j >> 8 {
				nonce = append(nonce, byte(j))
			}
			no := append(oprhash, nonce...)
			lx.Hash(no)
		}
	})
	b.Run("hash again", func(b *testing.B) {
		nonce := []byte{0, 0}
		for i := 0; i < b.N; i++ {
			nonce = nonce[:0]
			for j := i; j > 0; j = j >> 8 {
				nonce = append(nonce, byte(j))
			}
			no := append(oprhash, nonce...)
			lx.Hash(no)
		}
	})
}

func TestKnownHashes(t *testing.T) {

	known := map[string]string{
		"":       "66afa4d58ff4b99ef77f7bc2dc7567a23ccb47edab1486fccc3e9556bc64e9cc",
		"foo":    "2b72065417eaa9304b13521ce82effd183c934f015098ed8adaa9383f10c67c1,",
		"bar":    "eabbf4bc9279564ae33d398b88f7793c5cf0c9ce32ace0dcb0d9738543545419,,",
		"pegnet": "7cbf638ae44144a92b249fbd5b91ab25f438d60c21df5e8188d1d3c16e462b4a,,",
	}

	for k, v := range known {
		hash := lx.Hash([]byte(k))
		val, _ := hex.DecodeString(v)

		if !bytes.Equal(hash, val) {
			t.Errorf("mismatch for %s. got = %s, want = %s", k, hex.EncodeToString(hash), v)
		}
	}

}
