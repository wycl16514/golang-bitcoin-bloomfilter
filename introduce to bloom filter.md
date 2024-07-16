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


The merkleblock command is constructed by using the bloom filter, when asking a bitcoin full node to return a merkleblock message, the client need to send following info to the full node:

1, the number of buckets, the number is specified in bytes(1 byte equals to 8 buckets)

2, number of hash functions 

3, a "tweak", that is a number which is used to effect the result of hash functions, the aim of this field is to cause the result of hash function more random

4, bit field that results from the running bloom filter over the item of interest 

In order to easily create given number of hash functions, bitcoins use one hash function but set it to different seed everytime, it uses the murmur3 hash function which is time efficient and randomly
enough, and the seed used to change the behavior of the murmur3 funtion is given by following:

i * 0xfba4c795 + tweak

the i is the i-th hash function, and the value 0xfba4c795 is called BIT37_CONSTANT because it is defined by protocol of BIP37. Let's use code to simulate the above process:

```go
package main

import (
	"fmt"

	"github.com/spaolacci/murmur3"
)

const (
	BIP37_CONSTANT = 0xfba4c795
)

func main() {
	filedSize := 2
	functionNum := 2
	tweak := 42
	buckets := make([]byte, filedSize*8)
	strs := []string{"hello world", "goodbye"}
	for _, str := range strs {
		for i := 0; i < functionNum; i++ {
			seed := uint32(i*BIP37_CONSTANT + tweak)
			idx := murmur3.Sum32WithSeed([]byte(str), seed) % uint32(len(buckets))
			buckets[idx] = 1
		}
	}

	fmt.Printf("%x\n", buckets)
}

```

Running above code will give following result:
```go
01000000000000000000000001010100
```
There are 16 buckets, 4 of them have value 1, a string other than "hello world" and "goodbye" go through two hash functions, and map to 2 buckets, since the mapping is random enough, the probability of
first bucket has value 1 is 1/4, therefore 2 buckets both has value 1 is 1/16, if we have 160 strings, then we will have 10 collisions on average, then these 10 items will be the leaf node of
merkle tree. 

Now create a new folder named bloom-filter, add a new file named bloomfilter.go add code like following:
```go
package bloomfilter

import (
	"github.com/spaolacci/murmur3"
)

const (
	BIP37_CONSTANT = 0xfba4c795
)

type BloomFilter struct {
	size      uint64
	buckets   []byte
	funcCount uint64
	tweak     uint64
}

func NewBloomFilter(size uint64, funcCount uint64, tweak uint64) *BloomFilter {
	return &BloomFilter{
		size:    size,
                funcCount: funcCount,
		buckets: make([]byte, size*8),
		tweak:   tweak,
	}
}

func (b *BloomFilter) Add(item []byte) {
	for i := 0; i < int(b.funcCount); i++ {
		seed := uint32(uint64(i*BIP37_CONSTANT) + b.tweak)
		idx := murmur3.Sum64WithSeed(item, seed) % uint64(len(b.buckets))
		b.buckets[idx] = 1
	}
}

```

The item for the Add method will tell the full node what kind of transaction we are interesting, if we put a wallet address to the add, and send the filter to full node, full node will run through all
transactions, and put the wallet address of in the transaction to the filter, if the wallet address have all its buckets set to 1, then the given transaction will include in the merkleblock command.

Now question comes to how to send the filter to full node, we will use the filterload command, following is an example payload of the command:

0a4000600a080000010940050000006300000000

Let's put it into fields:

1, The first field is length of following data chunk, it is varint int, the value for given above example is 0x0a, actually it is the size field of BlommFilter

2, the following 10 bytes its value that we convert buckets from bits to bytes: 4000600a080000010940

3, the following 4 bytes is the number of hash functions, it is in little endian format: 05000000

4, the following 4 bytes in little endian format is value of tweak.

5, the last byte is called matched item flag: 00

Let's ad code to transfer the BloomFilter struct to filterload command:

```go
type FilterLoadMessage struct {
	payload []byte
}

func (f *FilterLoadMessage) Command() string {
	return "filterload"
}

func (f *FilterLoadMessage) Serialize() []byte {
	return f.payload
}

func (b *BloomFilter) BitsToBytes() []byte {
	if len(b.buckets)%8 != 0 {
		panic("length of buckets should divide over 8")
	}

	result := make([]byte, len(b.buckets)/8)
	for i, bit := range b.buckets {
		byteIndex := i / 8
		bitIndex := i % 8
		if bit == 1 {
			result[byteIndex] |= 1 << bitIndex
		}
	}

	return result
}

func (b *BloomFilter) FilterLoadMsg() *FilterLoadMessage {
	payload := make([]byte, 0)
	size := big.NewInt(int64(b.size))
	payload = append(payload, transaction.EncodeVarint(size)...)
	payload = append(payload, b.BitsToBytes()...)
	funcCount := big.NewInt(int64(b.funcCount))
	payload = append(payload, transaction.BigIntToLittleEndian(funcCount, transaction.LITTLE_ENDIAN_4_BYTES)...)
	tweak := big.NewInt(int64(b.tweak))
	payload = append(payload, transaction.BigIntToLittleEndian(tweak, transaction.LITTLE_ENDIAN_4_BYTES)...)
	//include all transaction that have collision
	payload = append(payload, 0x01)
	return &FilterLoadMessage{
		payload: payload,
	}
}
```
Now let's do some tests for the BlommFilter:
```go
package main

import (
	"bloomfilter"
	"fmt"
)

func main() {
	//testing add of bloom filter
	bf := bloomfilter.NewBloomFilter(10, 5, 99)
	bf.Add([]byte("Hello World"))
	fmt.Printf("%x\n", bf.BitsToBytes())

	bf.Add([]byte("Goodbye!"))
	fmt.Printf("%x\n", bf.BitsToBytes())

	//testing filterload
	fmt.Printf("%x\n", bf.FilterLoadMsg().Serialize())
}

```
Run the aboved code we will get the following result:
```go
0000000a080000000140
4000600a080000010940
0a4000600a080000010940050000006300000001
```







