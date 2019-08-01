# simMiner

The simulated miner allows you to run a test on hardware to evaluate speed, power, and hash rate of either 
Sha256 or the LXHash on that platform.

Usage:

simMiner <hash> [bits]

<hash> is either Sha256 or LXHash
[bits] is optional, but will default to 30 bits (about 1GB).  Takes about 10 minutes to initalize the BitMap for 1GB
on most common hardware tested.  Fewer bits (25 is about 32 MB) is pretty fast.