# LXRHash
Lookup XoR hash

This is a simple hash algorithm that takes advantage of a lookup table of randomized sets of bytes.  This lookup table 
consists of any number of 256 byte tables combined and sorted in one large table.  We then index into this large 
table to effectively look through the entire combination of tables as we translate the source data into a hash.

All parameters are specified.  The size of the lookup table (in numbers of 256 byte tables), the seed used to shuffle
the lookup table, the number of rounds to shuffle the table, and the size of the resulting hash.

This hash function has some interesting qualities.  Very large lookup tables will blow the cache on pretty much any 
processor or computer architecture. The number of bytes in the resulting hash can be increased for more security.
  
The lookup 
table has equal numbers of every byte value, but has them randomized over the whole table.  When hashing, the bytes from 
the source data are used to build offsets and state that are used to create the hash.

In developing this hash, the goal was to produce very randomized hashes as outputs, with a strong avalanche response to 
any change to any source byte.

LRXHash was only developed this hash as a thought experiment, but which none the less has some interesting qualities.

* the lookup table can be any size, so making a version that is ASIC resistant is possible
* at lookup table sizes where processor caches work, LXRHash for 256 bits is slightly faster than Sha256, at least 
in my tests
* at small lookup table sizes, LXRHash would be very trivial to impliment as an ASIC, and would be very fast
* at large lookup table sizes, LXRHash is ASIC and GPU resistent.

While this hash may be reasonable for use as PoW in mining on an immutable ledger that provides its own security, 
not nearly enough testing has been done to use as a fundamental part in cryptography or security.  For fun, it 
would be cool to do such testing.

The actual implementation is very small.  Assuming the lookup table is fixed (w.maps), the implementation follows, 
but look to the source code for comments and commentary on the implementation:
```go
type LXRHash struct {
	ByteMap  []byte // Integer Offsets
	MapSize  int64  // Size of the translation table
	Passes   int    // Passes to generate the rand table
	Seed     int64  // An arbitrary number used to create the tables.
	HashSize uint32 // Number of bytes in the hash
}

func (w LXRHash) Hash(src []byte) []byte {
	hashes := make([]int64,w.HashSize)
	var offset = int64(len(src))
	var last1, last2, last3 int64

	v := byte(offset)
	var idx1, idx2 int64

	step := func(i int, v2 int64) {
		offset = last1<<7 ^ last2<<3 ^ last3<<9 ^ offset<<8 ^ offset>>1 ^ idx2 ^ int64(v)
		idx1 = int64(uint64(idx1^offset^v2) % uint64(w.MapSize))
		v = w.ByteMap[idx1] ^ v
		last3 = last2>>2 ^ last3
		last2 = last1<<3 ^ last2
		last1 = int64(v) ^ v2 ^ last1<<1
		idx2 = idx2 ^ hashes[uint32(i)%w.HashSize]
	}

	// Pass through the source bytes, building up lastX values, hashes[], and offset
	for i, v2 := range src {
		step(i, int64(v2))
		hashes[uint32(i)%w.HashSize] = last3 ^ int64(v^v2) ^ idx2
	}

	// Reduction pass
	bytes := make([]byte, w.HashSize)
	for i, h := range hashes {
		step(i, h)
		idx2 := int64(uint64(int64(v)^offset) % uint64(w.MapSize))
		bytes[i] = w.ByteMap[idx2]
	}

	return bytes
}

```

The generation of the lookup table:
```go
const (
	firstrand = int64(2458719153079158768)
	firstb    = int64(4631534797403582785)
	firstv    = int64(3523455478921636871)
)
func (w *LXRHash) Init(Seed, MapSize int64, HashSize, Passes int) {
	MapSize = MapSize * 256 // Ensure the mapsize is a multiple of 256
	w.ByteMap = make([]byte, int(MapSize))
	w.HashSize = uint32(HashSize)
	w.MapSize = MapSize
	w.Seed = Seed
	w.Passes = Passes
	w.ReadTable()
}
// ReadTable
func (w *LXRHash) ReadTable() {
	filename := fmt.Sprintf("lrx%d.%d.%x.%x.dat", w.HashSize*8, w.Passes, w.Seed, w.MapSize)
	// Try and load our byte map.
	dat, err := ioutil.ReadFile(filename)

	// If loading fails, or it is the wrong size, generate it.  Otherwise just use it.
	if err != nil || len(dat) != int(w.MapSize) {
		w.GenerateTable()
		w.WriteTable(filename)
	} else {
		copy(w.ByteMap[:int(w.MapSize)], dat)
	}
}
// WriteTable
func (w *LXRHash) WriteTable(filename string) {
	os.Remove(filename)
	fo, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()
	if _, err := fo.Write(w.ByteMap[:]); err != nil {
		panic(err)
	}
}
// GenerateTable
func (w *LXRHash) GenerateTable() {
	offset := w.Seed ^ firstrand
	b := w.Seed ^ firstb
	v := int64(firstv)
	rand := func(i int64) int64 {
		offset = offset<<9 ^ offset>>1 ^ offset>>7 ^ b
		v = int64(w.ByteMap[uint64(offset^b)%uint64(w.MapSize)]) ^ v<<8 ^ v>>1
		b = v<<7 ^ v<<13 ^ v<<33 ^ v<<52 ^ b
		return int64(uint64(offset) % uint64(w.MapSize))
	}
	for i := range w.ByteMap {
		w.ByteMap[i] = byte(i)
	}
	for loops := 0; loops < w.Passes; loops++ {
		fmt.Println("Pass ", loops)
		for i := range w.ByteMap {
			j := rand(int64(i))
			w.ByteMap[i], w.ByteMap[j] = w.ByteMap[j], w.ByteMap[i]
		}
	}
}
```
