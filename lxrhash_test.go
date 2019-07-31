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

func TestKnownHashes(t *testing.T) {

	known := map[string]string{
		"":       "66afa4d58ff4b99ef77f7bc2dc7567a23ccb47edab1486fccc3e9556bc64e9cc",
		"foo":    "7dda54f8d5efcd6928870bdc9ece900b320e897bce4814e9010cc08647c197ae",
		"bar":    "fe2cb7f3cef5702a1cb4712434085afe1efdef1d2563291e4883cd2a3ea1e074",
		"pegnet": "cd45b08c0619d78e2a810c4e6462296ec51ae4fd0f73a54a154a97a54942297e",
	}

	for k, v := range known {
		hash := lx.Hash([]byte(k))
		val, _ := hex.DecodeString(v)

		if bytes.Compare(hash, val) != 0 {
			t.Errorf("mismatch for %s. got = %s, want = %s", k, hex.EncodeToString(hash), v)
		}
	}

}
