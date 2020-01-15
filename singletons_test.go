package lxr

import (
	"bytes"
	"fmt"
	"testing"
)

func testSize(t *testing.T, bits uint64, buf []byte, reference string) {
	one := Init(Seed, bits, HashSize, Passes)
	two := Init(Seed, bits, HashSize, Passes)

	defer Release(one)
	defer Release(two)

	if one != two {
		t.Errorf("[%d] two separate pointers for singleton: %x and %x", bits, &one, &two)
	}

	three := new(LXRHash)
	three.Init(Seed, bits, HashSize, Passes)

	if one == three {
		t.Errorf("[%d] separate instance is the same as singleton: %x and %x", bits, &one, &three)
	}

	res := fmt.Sprintf("%x", one.Hash(buf))
	if res != reference {
		t.Errorf("[%d] incorrect hash result, want = %s got = %s", bits, reference, res)
	}

	if !bytes.Equal(one.Hash(buf), two.Hash(buf)) {
		t.Errorf("[%d] one and two provided different hash results", bits)
	}
	if !bytes.Equal(one.Hash(buf), three.Hash(buf)) {
		t.Errorf("[%d] one and three provided different hash results", bits)
	}
}

func TestInit(t *testing.T) {
	buf := []byte("test string")

	testSize(t, 8, buf, "abab21b95cee68a5d70d871161e092530638b3b4bd4e88cadab3a5d6bbcf5f80")
	testSize(t, 9, buf, "c56d9652bc709713af194e2a64e0e2ff1bc0980b2395c772187186fbbf6cf9a9")
	testSize(t, 10, buf, "3c7df3fac7481d630571926c9be01056b36822baccdf2b1872936194477df2ff")
}

func TestRelease(t *testing.T) {
	one := Init(Seed, 8, HashSize, Passes)
	oneB := Init(Seed, 8, HashSize, Passes)

	Release(one)

	two := Init(Seed, 8, HashSize, Passes)

	if one != two {
		t.Errorf("released the instance despite another reference")
	}

	Release(oneB)
	Release(two)

	oneB = Init(Seed, 8, HashSize, Passes)

	if one == oneB {
		t.Errorf("singleton wasn't released after all references destroyed")
	}

	buf := []byte("test string")
	res := fmt.Sprintf("%x", one.Hash(buf))

	if res != "abab21b95cee68a5d70d871161e092530638b3b4bd4e88cadab3a5d6bbcf5f80" {
		t.Errorf("original singleton was destroyed during release")
	}
}
