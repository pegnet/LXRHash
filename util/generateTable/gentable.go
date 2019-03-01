package main

import (
	"fmt"
	"github.com/pegnet/LXR256"
)

func main() {
	lx := new(lxr.LXRHash)
	lx.GenerateTable()
	fmt.Println("byteMap := []byte {")
	for i, b := range lx.ByteMap {
		fmt.Printf("0x%02x, ", b)
		if (i+9)%16 == 0 {
		} else if (i+9)%8 == 0 {
			fmt.Print(" ")
			fmt.Println()
		}
	}
	fmt.Println("}")
}
