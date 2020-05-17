## Erasure coding constrains
* The number of data/parity shards must match the numbers used for encoding. 
* The order of shards must be the same as used when encoding. 
* You may only supply data you know is valid. 
* Invalid shards should be set to nil. 

So we need the order of data when getting Echo message, can we get the order from 