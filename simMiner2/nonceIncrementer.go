package main

type NonceIncrementer struct {
	Nonce         []byte
	lastNonceByte int
}

func NewNonceIncrementer(id int) *NonceIncrementer {
	n := new(NonceIncrementer)
	n.Nonce = []byte{byte(id), 0}
	n.lastNonceByte = 1
	return n
}

// NextNonce is just counting to get the next nonce. We preserve
// the first byte, as that is our ID and give us our nonce space
//	So []byte(ID, 255) -> []byte(ID, 1, 0) -> []byte(ID, 1, 1)
func (i *NonceIncrementer) NextNonce() {
	idx := len(i.Nonce) - 1
	for {
		i.Nonce[idx]++
		if i.Nonce[idx] == 0 {
			idx--
			if idx == 0 { // This is my prefix, don't touch it!
				rest := append([]byte{1}, i.Nonce[1:]...)
				i.Nonce = append([]byte{i.Nonce[0]}, rest...)
				break
			}
		} else {
			break
		}
	}

}
