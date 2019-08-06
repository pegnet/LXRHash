# LXRHash
Lookup XoR Hash:  First RAM Hash Algorithm
---------
LXRHash uses an XOR/Shift random number generator coupled with a lookup table of randomized sets of bytes.  
The lookup table consists of any number of 256 byte tables combined and shuffled into one large byte lookup table.  We 
then index into this large table to translate the state built up while hashing into deterministic but random byte values.

Using a 1GB lookup table results in a RAM Hash PoW algorithm that spends over 90% of its execution time waiting on memory (RAM) than it does computing the hash.  This means far less power consumption, and ASIC and GPU resistence.  The ideal platform for PoW using a RAM Hash is a Single Board Computer like a Raspberry PI 4 with 2GB of memory.

All parameters are specified.  The size of the lookup table is specified in bits for the index, the seed used to shuffle
the lookup table, the number of rounds to shuffle the table, and the size of the resulting hash.

Because the LXRHash is parameterized in this way, as computers get faster and larger memory caches, the LXRHash can be set to use 2GB or 16GB or more.  The Memory bottleneck to computation is much easier to manage than attempts to find computational algorithms that cannot be executed faster and cheaper with custom hardware, or specialty hardware like GPUs.

So to quickly interate over LXRHash features:  
* Very large lookup tables will blow the memory caches on pretty much any processor or computer architecture
* The size of the table can be increased to counter improvements in memory caching
* The number of bytes in the resulting hash can be increased for more security (greater hash space), without significantly
more processing time
* LXRHash *can* be fast by using small lookup tables
* ASIC implementations for small tables would be very easy and very fast
* LXRHash only uses iterators (for indexing) shifts, binary ANDs and XORs, and random byte lookups
* The use case for LXRHash is Proof of Work (PoW), not cryptographic hashing

The Lookup 
-------
The look up table has equal numbers of every byte value, and shuffled deterministically.  When hashing, the bytes 
from the source data are used to build offsets and state that are in turn used to map the next byte of source.

In developing this hash, the goal was to produce very randomized hashes as outputs, with a strong avalanche response to 
any change to any source byte.  This is the prime requirement of PoW.  Because of the limited time to perform hashing
in a blockchain, collision avoidence is important but not critical.  More critical is ensuring engineering the output 
of the hash isn't possible.

LRXHash was origionally developed as a thought experiment, yet the result yeilds some interesting qualities.

* the lookup table can be any size, so making a version that is ASIC resistant is possible by using very big lookup tables.  Such tables blow the processor caches on CPUs and GPUs, making the speed of the hash dependent on random access of memory, not processor power.  Using 1 GB lookup table, a very fast ASIC improving hashing is limited to about ~10% of the computational time for the hash.  90% of the time hashing isn't spent on computation but is spent waiting for 
memory access.  
* at smaller lookup table sizes where processor caches work, LXRHash can be modified to be very fast.
* LXRHash would be an easy ASIC design as it only uses counters, decrements, XORs, and shifts. 
* the hash is trivially altered by changing the size of the lookup table, the seed, size of the hash produced. Change any parameter and you change the space from which hashes are produced.
* The Microprocessor in most computer systems accounts for 10x the power requirements of memory.  If we consider PoW on a device over time, then LXRHash is estimated to reduce power requirements by about a factor of 10.

While this hash may be reasonable for use as PoW in mining on an immutable ledger that provides its own security, 
not nearly enough testing has been done to use as a fundamental part in cryptography or security.  For fun, it 
would be cool to do such testing.

## Testing
To run the LXRHash benchmark test:
```shell
cd testing
go test
```

