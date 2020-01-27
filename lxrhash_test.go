// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package lxr

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"math/rand"
	"testing"
)

var lx LXRHash
var oprhash []byte

func init() {
	lx.Init(0xfafaececfafaecec, 30, 256, 5)
	oprhash = lx.Hash([]byte(`Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nunc dapibus pretium urna, mollis aliquet elit cursus ac. Sed sodales, erat ut volutpat viverra, ante urna pretium est, non congue augue dui sed purus. Mauris vitae mollis metus. Fusce convallis faucibus tempor. Maecenas hendrerit, urna eu lobortis venenatis, neque leo consequat enim, nec placerat tellus eros quis diam. Donec quis vestibulum eros. Maecenas id vulputate justo. Quisque nec feugiat nisi, lacinia pulvinar felis. Pellentesque habitant sed.`))
}

func BenchmarkHash(b *testing.B) {
	normalHash := func(b *testing.B) {
		nonce := []byte{0, 0}
		for i := 0; i < b.N; i++ {
			nonce = nonce[:0]
			for j := i; j > 0; j = j >> 8 {
				nonce = append(nonce, byte(j))
			}
			no := append(oprhash, nonce...)
			lx.Hash(no)
		}
	}

	flatHash := func(b *testing.B) {
		nonce := []byte{0, 0}
		for i := 0; i < b.N; i++ {
			nonce = nonce[:0]
			for j := i; j > 0; j = j >> 8 {
				nonce = append(nonce, byte(j))
			}
			no := append(oprhash, nonce...)
			lx.FlatHash(no)
		}
	}

	batchHash := func(b *testing.B) {
		// Create sets based on b.N
		batchsize := 128 // Feel free to tweak
		sets := (b.N / batchsize) + 1
		batches := make([][][]byte, sets)
		for i := range batches {
			batches[i] = make([][]byte, batchsize)
			for j := range batches[i] {
				batches[i][j] = make([]byte, 4)
				binary.BigEndian.PutUint32(batches[i][j], uint32((i*batchsize)+j))
			}
		}

		// If you want to skip the setup, you can start the timer here.
		// But the setup is included in the others, so it is also being
		// included here. It might be better to make the batches on demand
		// vs upfront.
		for i := range batches {
			lx.HashParallel(oprhash, batches[i])
		}
	}

	// The last hashing function always runs faster for some reason.
	// So mix them up a bit
	b.Run("hash", normalHash)
	b.Run("flat hash", flatHash)
	b.Run("HashParallel", batchHash)

	b.Run("hash again", normalHash)
	b.Run("flat hash again", flatHash)
	b.Run("HashParallel again", batchHash)
}

func TestKnownHashes(t *testing.T) {

	known := map[string]string{
		"":       "66afa4d58ff4b99ef77f7bc2dc7567a23ccb47edab1486fccc3e9556bc64e9cc",
		"abcde":  "00e9ef8262f154b6aef3b4bb1a95644bbd651040df34c3d88dd696d519445989",
		"bar":    "66a7c02adcf00ed55a11877fa543ccc27a0a4c59268cc36cd8fe9616ce6cda63",
		"foo":    "93a2eaf76b8cc21610601fb5a87f8f6ea57ef0fc1e6eaf414e7b6eac186bca16",
		"pegnet": "84c5bc3b47965e0fff9e66871b94dd7d2cd1f866102a6c1cd7ef30eb3ee737ef",

		"0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000": "e169f393b60ef4e74fa2b3f514451523911a3c9929c76b39bd46f448979e784f",

		"1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000": "da715b359c07e94c3db8e7ca0fb2786ffc1d40cae2d02d4d193da4c5f0b28e6c",
		"2000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000": "fe788f9bb86a3b014f1b7b5247bee1f88471a795f17d3d8d9555a2d74dd56a66",
		"3000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000": "3122704067ec22284d47f8ed30e2e218bab4b9885c951f5578ae958ea88d2242",
		"4000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000": "81d4ba04b98fa2d9af34af88323904be70c0dc47bd4cbf5d5ba39ff684a41cf0",
		"5000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000": "aece6cc62f94ea08c7289d52caeee7d239efecfc72fac11b78bee157675939f5",

		"0000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000": "f84af0f18a9be4d89194b658027ba2e4d55ec0d6ad681ba6667e43f27c1cbf63",
		"0000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000": "0b526f63210add8d7984bbd0ef1cffd2e3fc263a1fd548bdb8a4e33b7838e8c4",
		"0000000000000000000000000000000000000000000000003000000000000000000000000000000000000000000000000000000": "434395f5efa71773c2b4b2f4c0fd9d5a88b2010002080fa54cb4a8163bcb827c",
		"0000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000": "42372a30bbc752c654e072b06d680ad77357caf87353a0f4e3e012158fa6928f",
		"0000000000000000000000000000000000000000000000005000000000000000000000000000000000000000000000000000000": "136187976bca29b0d77ca8f29846e81e3f6111dcf016f5f0e78bd912db6180e1",

		"0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001": "2079f06de6d91efa953667e16fdfb573f2d0196c0d5ffd7f3a27243497a26a33",
		"0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002": "b4ed552867c41fcc73190374b38188a424f014d906d2d8603bc68995fcee82da",
		"0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003": "89d32d663342ef54d27ce87ee1da784c239921954393a083c63564fd4be98f57",
		"0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000004": "211dc5cbe8003e7c992f4d82788c9bb76cd69d7623cb78b1454266b57248852e",
	}

	for k, v := range known {
		hash := lx.Hash([]byte(k))
		val, _ := hex.DecodeString(v)

		if !bytes.Equal(hash, val) {
			t.Errorf("mismatch for %s. got = %s, want = %s", k, hex.EncodeToString(hash), v)
		}
	}

	for k, v := range known {
		hash := lx.FlatHash([]byte(k))
		val, _ := hex.DecodeString(v)

		if !bytes.Equal(hash, val) {
			t.Errorf("flathash mismatch for %s. got = %s, want = %s", k, hex.EncodeToString(hash), v)
		}
	}

	for k, v := range known {
		val, _ := hex.DecodeString(v)

		res := lx.HashParallel([]byte(k), [][]byte{[]byte{}})
		if len(res) != 1 {
			t.Error("missing results")
			t.FailNow()
		}
		if !bytes.Equal(res[0], val) {
			t.Errorf("HashParallel (zero-batch) mismatch for %s. got = %s, want = %s", k, hex.EncodeToString(res[0]), v)
		}
	}

	for k, v := range known {
		val, _ := hex.DecodeString(v)

		res := lx.HashParallel([]byte{}, [][]byte{[]byte(k)})
		if len(res) != 1 {
			t.Error("missing results")
			t.FailNow()
		}
		if !bytes.Equal(res[0], val) {
			t.Errorf("HashParallel (zero-base) mismatch for %s. got = %s, want = %s", k, hex.EncodeToString(res[0]), v)
		}
	}

}

func TestLXRHash_Hash(t *testing.T) {
	for i := 0; i < 1000; i++ {
		data := make([]byte, rand.Intn(100))
		h1, h2 := lx.Hash(data), lx.FlatHash(data)
		if bytes.Compare(h1, h2) != 0 {
			t.Errorf("mismatch hashes\n%x\n%x", h1, h2)
		}
	}
}

func TestBatch(t *testing.T) {
	batchsize := 512
	batch := make([][]byte, batchsize)
	start := uint32(0)
	static := make([]byte, 32)
	rand.Seed(0) // I want this deterministic
	rand.Read(static)

	for i := range batch {
		batch[i] = make([]byte, 4)
		binary.BigEndian.PutUint32(batch[i], start+uint32(i))
	}

	results := lx.HashParallel(static, batch)
	for i := range results {
		// do something with the result here
		// nonce = batch[i]
		// input = append(base, batch[i]...)
		// hash = results[i]
		h := results[i]
		h2 := lx.Hash(append(static, batch[i]...))
		h3 := lx.FlatHash(append(static, batch[i]...))

		if !bytes.Equal(h, h2) {
			t.Errorf("not same, batch failed\n%x\n%x", h, h2)
		}
		if !bytes.Equal(h, h3) {
			t.Errorf("not same, batch failed\n%x\n%x", h, h3)
		}
	}
}

func TestAbortSettings(t *testing.T) {
	if b, v := AbortSettings(0xffac55c69ecabf4f); b != 1 || v != 0xac {
		t.Errorf("unexpected")
	}
}

func target(b []byte) uint64 {
	return uint64(b[7]) | uint64(b[6])<<8 | uint64(b[5])<<16 | uint64(b[4])<<24 |
		uint64(b[3])<<32 | uint64(b[2])<<40 | uint64(b[1])<<48 | uint64(b[0])<<56
}
