# LXRHash
Lookup XoR hash
---------
This is a simple hash algorithm that takes advantage of a lookup table of randomized sets of bytes.  This lookup table 
consists of any number of 256 byte tables combined and sorted in one large table.  We then index into this large 
table to effectively look through the entire combination of tables as we translate the source data into a hash.

All parameters are specified.  The size of the lookup table (in numbers of 256 byte tables), the seed used to shuffle
the lookup table, the number of rounds to shuffle the table, and the size of the resulting hash.

This hash function has some interesting qualities.  Very large lookup tables will blow the cache on pretty much any 
processor or computer architecture. The number of bytes in the resulting hash can be increased for more security, without any more processing time.  Note, while this approach *can* be fast, this implemenation isn't.  The use case is aimed at Proof of Work (PoW), not cryptographic hashing.
  
The lookup 
-------
table has equal numbers of every byte value, but has them randomized over the whole table.  When hashing, the bytes from 
the source data are used to build offsets and state that are used to create the hash.

In developing this hash, the goal was to produce very randomized hashes as outputs, with a strong avalanche response to 
any change to any source byte.

LRXHash was only developed this hash as a thought experiment, and yeilds some interesting qualities.

* the lookup table can be any size, so making a version that is ASIC resistant is possible by using very big lookup tables.  Such tables blow the processor caches on CPUs and GPUs, making the speed of the hash dependent on random access of memory, not processor power.  Using 1 GB lookup table, a very fast ASIC improving hashing is limited to about ~1/3 of the computational time for the hash.  2/3 of the time is spent waiting for memory access.  
* at smaller lookup table sizes where processor caches work, LXRHash can be modified to be very fast.
* LXRHash would be an easy ASIC design as it only uses counters, decrements, XORs, and shifts. 
* the hash is trivially altered by changing the size of the lookup table, the seed, size of the hash produced. Change any parameter and you change the space from which hashes are produced.

While this hash may be reasonable for use as PoW in mining on an immutable ledger that provides its own security, 
not nearly enough testing has been done to use as a fundamental part in cryptography or security.  For fun, it 
would be cool to do such testing.

The actual implementation is very small, and see the code for the most accurate source. The code is presented here without comments to illustrate its small size (Hash() is ~ 26 lines of go) :
```go
type LXRHash struct {
	ByteMap     []byte // Integer Offsets
	MapSize     uint64 // Size of the translation table
	MapSizeBits uint64 // Size of the ByteMap in Bits
	Passes      uint64 // Passes to generate the rand table
	Seed        uint64 // An arbitrary number used to create the tables.
	HashSize    uint64 // Number of bytes in the hash
}
func (w LXRHash) Hash(src []byte) []byte {
	hashes := make([]uint64, w.HashSize)
	var lastStage = w.Seed
	var stages, stages2 [11]uint64
	MapMask := w.MapSize - 1
	step := func(i uint64, v2 uint64) {
		stages[0] = stages[0] ^ lastStage ^ v2 ^ uint64(w.ByteMap[(lastStage^v2<<9)%w.MapSize])<<4
		for i := len(stages) - 1; i >= 0; i-- {
			stage := stages[i]
			if i > 0 {
				stages[i] = stages[i-1]<<7 ^ stages[i-1]>>1 ^ stage ^ uint64(w.ByteMap[(stage^v2<<9)&MapMask])<<16
				lastStage = stage ^ lastStage<<11 ^ lastStage>>1
			}
		}
		stages, stages2 = stages2, stages
	}
	for i, v2 := range src {
		idx := uint64(i)
		step(idx, uint64(v2))
		hash := hashes[idx%w.HashSize]
		hashes[idx%w.HashSize] = lastStage ^ hash<<21 ^ hash>>1
	}
	bytes := make([]byte, w.HashSize)
	for i, h := range hashes {
		step(uint64(i), h)
		idx2 := (stages[0] ^ lastStage) & MapMask
		bytes[i] = w.ByteMap[idx2]
	}
	return bytes
}


```

The generation of the lookup table:
```go
const (
	firstrand = uint64(2458719153079158768)
	firstb    = uint64(4631534797403582785)
	firstv    = uint64(3523455478921636871)
)
func (w *LXRHash) Init(Seed, MapSizeBits, HashSize, Passes uint64) {
	if MapSizeBits < 8 {
		panic(fmt.Sprintf("Bad Map Size in Bits.  Must be between 8 and 34 bits, was %d", MapSizeBits))
	}
	MapSize := uint64(1) << MapSizeBits
	w.ByteMap = make([]byte, int(MapSize))
	w.HashSize = (HashSize + 7) / 8
	w.MapSize = MapSize
	w.MapSizeBits = MapSizeBits
	w.Seed = Seed
	w.Passes = Passes
	w.ReadTable()
}
func (w *LXRHash) ReadTable() {
	filename := fmt.Sprintf("lrxhash.seed-%x.passes-%d.size-%d.dat", w.Seed, w.Passes, w.MapSizeBits)
	println("Reading ByteMap Table ", filename)
	dat, err := ioutil.ReadFile(filename)
	if err != nil || len(dat) != int(w.MapSize) {
		println("Table not found, Generating ByteMap Table ")
		w.GenerateTable()
		fmt.Println("writeing ByteMap Table ")
		w.WriteTable(filename)
		fmt.Println("Done")
	} else {
		w.ByteMap = dat
	}
}
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
func (w *LXRHash) GenerateTable() {
	offset := w.Seed ^ firstrand
	b := w.Seed ^ firstb
	v := firstv
	MapMask := w.MapSize - 1
	rand := func(i uint64) int64 {
		offset = offset<<9 ^ offset>>1 ^ offset>>7 ^ b
		v = uint64(w.ByteMap[(offset^b)&MapMask]) ^ v<<8 ^ v>>1
		b = v<<7 ^ v<<13 ^ v<<33 ^ v<<52 ^ b<<9 ^ b>>1
		return int64(uint64(offset) % uint64(w.MapSize))
	}
	start := time.Now().Unix()
	period := start
	println("Initalize the Table")
	for i := range w.ByteMap {
		if (i+1)%1000 == 0 && time.Now().Unix()-period > 10 {
			println(" Index ", i+1, " of ", len(w.ByteMap))
			period = time.Now().Unix()
		}
		w.ByteMap[i] = byte(i)
	}
	println("Shuffling the Table")
	for loops := 0; loops < int(w.Passes); loops++ {
		fmt.Println("Pass ", loops)
		for i := range w.ByteMap {
			if (i+1)%1000 == 0 && time.Now().Unix()-period > 10 {
				fmt.Printf(" Index %10d Meg of %10d Meg -- Pass is %5.1f%% Complete\n", i/1024000, len(w.ByteMap)/1024000, 100*float64(i)/float64(len(w.ByteMap)))
				period = time.Now().Unix()
			}
			j := rand(uint64(i))
			w.ByteMap[i], w.ByteMap[j] = w.ByteMap[j], w.ByteMap[i]
		}
		fmt.Printf(" Index %10d Meg of %10d Meg -- Pass is %5.1f%% Complete\n", len(w.ByteMap)/1024000, len(w.ByteMap)/1024000, float64(100))
	}
}
```
