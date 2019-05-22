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
	ByteMap  []byte // Integer Offsets
	MapSize  int64  // Size of the translation table
	Passes   int    // Passes to generate the rand table
	Seed     int64  // An arbitrary number used to create the tables.
	HashSize uint32 // Number of bytes in the hash
}
// Hash()
func (w LXRHash) Hash(src []byte) []byte {
	hashes := make([]int64, w.HashSize)
	var lastStage = int64(len(src)) ^ w.Seed ^ int64(w.HashSize)
	var stages [5]int64
	v := w.ByteMap[lastStage%w.MapSize]
	step := func(i int, v2 int64) int64 {
		lastStage = v2 ^ int64(i)<<16 ^ lastStage
		for i, stage := range stages {
			ui := uint64(i)
			stages[i] = stage<<(8+ui) ^ stage>>(1+ui) ^ lastStage
			lastStage = stage ^ lastStage<<5 ^ int64(v<<ui)
		}
		v = w.ByteMap[uint64(lastStage)%uint64(w.MapSize)] ^ v
		return hashes[uint32(i)%w.HashSize]
	}
	for i, v2 := range src {
		step(i, int64(v2))
		hashes[uint32(i)%w.HashSize] = lastStage ^
			int64(v^w.ByteMap[uint64(stages[0]+int64(v2))%uint64(w.MapSize)])
	}
	bytes := make([]byte, w.HashSize)
	for i, h := range hashes {
		step(i, h)
		idx2 := int64(uint64(int64(v)^lastStage) % uint64(w.MapSize))
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
	if MapSize&0xFF > 0 {
		panic(fmt.Sprintf("MapSize specified is not a multiple of 256, off by %d... Add %d to fix", 
		MapSize&0xFF, 256-MapSize&0xff))
	}
	w.ByteMap = make([]byte, int(MapSize))
	w.HashSize = uint32(HashSize) / 8
	w.MapSize = MapSize
	w.Seed = Seed
	w.Passes = Passes
	w.ReadTable()
}
// ReadTable
func (w *LXRHash) ReadTable() {
	filename := fmt.Sprintf("lrxhash.seed-%x.passes-%d.size-%d.dat", w.Seed, w.Passes, w.MapSize)
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
	for loops := 0; loops < w.Passes; loops++ {
		fmt.Println("Pass ", loops)
		for i := range w.ByteMap {
			if (i+1)%1000 == 0 && time.Now().Unix()-period > 10 {
				fmt.Printf(" Index %10d Meg of %10d Meg -- Pass is %5.1f%% Complete\n", 
				    i/1024000, len(w.ByteMap)/1024000, 100*float64(i)/float64(len(w.ByteMap)))
				period = time.Now().Unix()
			}
			j := rand(int64(i))
			w.ByteMap[i], w.ByteMap[j] = w.ByteMap[j], w.ByteMap[i]
		}
		fmt.Printf(" Index %10d Meg of %10d Meg -- Pass is %5.1f%% Complete\n", 
		    len(w.ByteMap)/1024000, len(w.ByteMap)/1024000, float64(100))
	}
}

```
