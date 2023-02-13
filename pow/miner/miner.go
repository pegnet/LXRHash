package main

import (
	"github.com/pegnet/LXRHash/pow"
)

func main() {
	for i := uint64(8); i < 31; i++ {
		new(pow.LxrPow).Init(i, 6)
	}
}
