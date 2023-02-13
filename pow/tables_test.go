// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package pow

import "testing"

func TestLxrPow_Init(t *testing.T) {
	for i:= uint64(8); i<30; i++ {
		p := new(LxrPow)
		p.Init(i,6)
	}
}
