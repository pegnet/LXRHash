// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package testing_test

import (
	"fmt"
	"testing"

	lxr "github.com/pegnet/LXRHash"
)

func TestBitsOfHash(t *testing.T) {
	LX.Init(Seed, MapSizeBits, HashSize, Passes)

	SumBitsChanged := 0
	TotalBits := 0
	bytesChanged := 0
	TotalBytes := 0

	SumBitsChangedV := 0
	TotalBitsV := 0
	bytesChangedV := 0
	TotalBytesV := 0

	const bufferlen = 40

	for i := 0; i < 10000; i++ {
		buf := Getbuf(40)
		var last []byte

		for j := 0; j < bufferlen; j++ {
			for k := uint(0); k < 8; k++ {
				bit := byte(1 << k)
				buf[j] = buf[j] ^ bit
				wv, _ := LX.HashValidate(buf, nil)
				if last != nil {
					for bc := 0; bc < 32; bc++ {
						for bcb := uint(0); bcb < 8; bcb++ {
							bit := byte(1) << bcb
							if wv[bc]&bit != last[bc]&bit {
								SumBitsChanged++
							}
							TotalBits++
						}
						if wv[bc] != last[bc] {
							bytesChanged++
						}
						TotalBytes++
					}
					for bc := 32; bc < 256-32; bc++ {
						for bcb := uint(0); bcb < 8; bcb++ {
							bit := byte(1) << bcb
							if wv[bc]&bit != last[bc]&bit {
								SumBitsChangedV++
							}
							TotalBitsV++
						}
						if wv[bc] != last[bc] {
							bytesChangedV++
						}
						TotalBytesV++
					}
				}
				buf[j] = buf[j] ^ bit
				last = wv
			}
		}
	}

	fmt.Println("hash")
	fmt.Printf("Total Bits  Changed %15d Total Bits  %15d Percent Changed %6.3f%% \n", SumBitsChanged, TotalBits, float64(SumBitsChanged)/float64(TotalBits)*100)
	fmt.Printf("Total Bytes Changed %15d Total Bytes %15d Percent Changed %6.3f%% \n", bytesChanged, TotalBytes, float64(bytesChanged)/float64(TotalBytes)*100)
	fmt.Println("Validation bytes")
	fmt.Printf("Total Bits  Changed %15d Total Bits  %15d Percent Changed %6.3f%% \n", SumBitsChangedV, TotalBitsV, float64(SumBitsChangedV)/float64(TotalBitsV)*100)
	fmt.Printf("Total Bytes Changed %15d Total Bytes %15d Percent Changed %6.3f%% \n", bytesChangedV, TotalBytesV, float64(bytesChangedV)/float64(TotalBytesV)*100)
	fmt.Printf(" 1 out of 256 %10.5f\n", float64(1)/2/2/2/2/2/2/2/2)
}

// Tests to ensure a single bit flipped in the Verification bytes results in the order of 50% of the hash
// changing.
func TestBitsOfValidation(t *testing.T) {
	LX.Init(Seed, MapSizeBits, HashSize, Passes)

	SumBitsChanged := 0
	TotalBits := 0
	bytesChanged := 0
	TotalBytes := 0

	const bufferlen = 40

	for i := 0; i < 10000; i++ {
		buf := Getbuf(40)

		for j := 0; j < 256-32; j++ {
			for k := uint(0); k < 8; k++ {
				bit := byte(1 << k)
				wvo, _ := LX.HashValidate(buf, nil)
				wvo[32+j] = wvo[32+j] ^ bit
				wvn, err := LX.HashValidate(buf, wvo)
				if err == nil {
					t.Fail()
				}
				for bc := 0; bc < 32; bc++ {
					for bcb := uint(0); bcb < 8; bcb++ {
						bit := byte(1) << bcb
						if wvo[bc]&bit != wvn[bc]&bit {
							SumBitsChanged++
						}
						TotalBits++
					}
					if wvo[bc] != wvn[bc] {
						bytesChanged++
					}
					TotalBytes++
				}
			}
		}
	}

	fmt.Println("hash")
	fmt.Printf("Total Bits  Changed %15d Total Bits  %15d Percent Changed %6.3f%% \n", SumBitsChanged, TotalBits, float64(SumBitsChanged)/float64(TotalBits)*100)
	fmt.Printf("Total Bytes Changed %15d Total Bytes %15d Percent Changed %6.3f%% \n", bytesChanged, TotalBytes, float64(bytesChanged)/float64(TotalBytes)*100)
	fmt.Println("Validation bytes")
	fmt.Printf(" 1 out of 256 %10.5f\n", float64(1)/2/2/2/2/2/2/2/2)
}

// Tests to ensure a single bit flipped in the Verification bytes results in the order of 50% of the hash
// changing.
func TestVerificationBytes(t *testing.T) {
	LX.Init(Seed, MapSizeBits, HashSize, Passes)

	LX.ValidationSize = 100000000
	LX.ValidationIndex = 30000000

	LXs := new(lxr.LXRHash2)
	LXs.Init(Seed, 10, 32, 5)

	SumBitsChanged := 0
	TotalBits := 0
	bytesChanged := 0
	TotalBytes := 0

	const bufferlen = 40

	for i := 0; i < 10000; i++ {
		buf := Getbuf(40)

		for j := 0; j < 256-32; j++ {
			for k := uint(0); k < 8; k++ {
				bit := byte(1 << k)
				wvo, _ := LX.HashValidate(buf, nil)
				wvo[32+j] = wvo[32+j] ^ bit
				wvn, err := LX.HashValidate(buf, wvo)
				if err == nil {
					t.Fail()
				}
				for bc := 0; bc < 32; bc++ {
					for bcb := uint(0); bcb < 8; bcb++ {
						bit := byte(1) << bcb
						if wvo[bc]&bit != wvn[bc]&bit {
							SumBitsChanged++
						}
						TotalBits++
					}
					if wvo[bc] != wvn[bc] {
						bytesChanged++
					}
					TotalBytes++
				}
			}
		}
	}

	fmt.Println("hash")
	fmt.Printf("Total Bits  Changed %15d Total Bits  %15d Percent Changed %6.3f%% \n", SumBitsChanged, TotalBits, float64(SumBitsChanged)/float64(TotalBits)*100)
	fmt.Printf("Total Bytes Changed %15d Total Bytes %15d Percent Changed %6.3f%% \n", bytesChanged, TotalBytes, float64(bytesChanged)/float64(TotalBytes)*100)
	fmt.Println("Validation bytes")
	fmt.Printf(" 1 out of 256 %10.5f\n", float64(1)/2/2/2/2/2/2/2/2)
}

// TestFastValidation(t *testing.T)
// Test Fast Validation
func TestFastValidation(t *testing.T) {
	LX.Init(Seed, 30, HashSize, Passes)

	LXs := new(lxr.LXRHash2)
	LXs.Init(Seed, 10, 256, 5)

	var VSize, FailCaught, Total [49]int

	for i := 0; i < 10000; i++ {

		for j := float64(1); j < 10; j++ {
			buf := Getbuf(40)
			percent := j / 100
			VSize[int(j-1)] = int(float64(LX.MapSize) * percent)
			LX.ValidationSize = uint64(VSize[int(j-1)])
			LX.ValidationIndex = 0
			wvo, _ := LXs.HashValidate(buf, nil)
			_, err := LX.HashValidate(buf, wvo)
			if err != nil {
				FailCaught[int(j)-1]++
			}
			Total[int(j)-1]++
		}

	}
	for i, fc := range FailCaught {
		fmt.Printf("Percent Backing %5.0f%%  Fail %15.13f%%\n",
			float64(VSize[i])/float64(LX.MapSize)*100,
			float64(fc)/float64(Total[i])*100)

	}

}
