In previous section, we ask our fullnode peer to return a merbkleblock command that we can verify whether given transactions of interested(the green boxes) are include in a block or not. And most of time we or the client dosen't want the fullnode know which transactions are interested to us, therefore we want to hide our target transactions in a group of transactions(The leafs of the merkle tree). 
Then we need an effective method to transfer info about that group of transactions to fullnode.That's where the data structure and algorithm of bloom filter comes into play.

![image](https://github.com/wycl16514/golang-bitcoin-core-merkle-tree/assets/7506958/8e79a944-601e-4845-be1e-00c4246c6d2f)

As the image shows, we are interesting the green boxes , but we want to hide our intention and we hide these two in a group of transactions that are H(A) to H(P), then we ask full node do you have info about all transaction about H(A) to H(P), then the full node will construct the merkleblock command to us, and we can compute the value of the root to verify whether our target transactions that are H(K) and H(N) are put onto the mainet or not.

There are 1.041 billion of transactions for bitcoin blockchain now, how can we quickly select the given several transactions out from 1 billion? That's where bloom filter come into play. Bloom filter is
a kind of data structure used for big data, think about sipder of google crawling web pages, given a url how the spider konw whehter this page is alread saved on the server of google or not. The way of 
doing this is, given the url, we have a group of buckets that are made up of bits, and we have a group of hash functions, the hash functions will hash the given string to the index of a given bucket,
each time we using a hash function to hash the url to a given bucket, we check the value of the bucket, if there is one bucket with value of 0, then the given page with the url is not saved before,
if all the bucket we visited have the value 1, then we are sure the page is alread saved on the server.

The logic of bloom filter as following:
![image](https://github.com/user-attachments/assets/30d0a10e-bd8c-45b0-9e53-353d914d262c)

There is a possiblity of fasle-positive for bloom filter, a given page may not saved on the server but the bloom filter give a positive result which means given the url that its page is not saved on the 
server, but all the buckets we visited have the value of 1. The posibility of false-positive can be leverage by enlarging the number of buckets, the more buckets you have, the less likly you will have 
a false positive.
create a new folder called bloom-filter, in main.go, we will create 10 buckets and using one hash function to hash a given string into a bucket:
```go
package main

import (
	ecc "elliptic_curve"
	"fmt"
	"math/big"
)

func main() {
	buckets := make([]byte, 10)
	h := ecc.Hash256("a random string")
	val := new(big.Int)
	val.SetBytes(h)
	fmt.Printf("hash value:%d\n", val.Int64())
	var opMod big.Int
	idx := opMod.Mod(val, big.NewInt(int64(10)))
	//set the value of given buckt to 1
	buckets[idx.Int64()] = 1
	fmt.Printf("%x\n", buckets)
}
```
Running aboved code we have the following result:
```go
hash value:3457011286314174106
00000000000000000100
```

To prevent any collision, for example if two different strings map to the same bucket, then we may have false-positive, in order to low the liklyhood of collision we can use more hash functions to map
a given object to more buckets, the exmaple code will like following:

```go
func main() {
	buckets := make([]byte, 10)
	strs := make([]string, 0)
	strs = append(strs, "hello world")
	strs = append(strs, "goodbyte")
	for _, str := range strs {
		//using two hash function to map given string to two buckets
		h := ecc.Hash256(str)
		val := new(big.Int)
		val.SetBytes(h)
		var opMod big.Int
		idx := opMod.Mod(val, big.NewInt(int64(10)))
		//set the value of given buckt to 1
		buckets[idx.Int64()] = 1

		h = ecc.Hash160([]byte(str))
		val.SetBytes(h)
		idx = opMod.Mod(val, big.NewInt(int64(10)))
		//set the value of given buckt to 1
		buckets[idx.Int64()] = 1
	}

	fmt.Printf("%x\n", buckets)
}

```
The result for running aboved code is :
```go
01010000000001000001
```






