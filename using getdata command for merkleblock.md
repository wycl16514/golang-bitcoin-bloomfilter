In previous section, we have bloom filter and create filterload command to send info about the filter to the full node. We still need another command name getdata to request filtered
block from the full node, a filtered block is asking full node to throw transactions to the filter we sent to it and include any transactions that can be matched by the filter(all 
buckets have value 1), then put all thosed filtered transactions into merkleblock command.

Let's have a look at the payload of getdata and put it into fields:

020300000030eb2540c41025690160a1014c577061596e32e426b712c7ca00000000000000030000001049847939585b0652fba793661c361223446b6fc41089b8be00000000000000

1, At the beginning its varint, the value in aboved data is 0x2 then we only need to get one byte.

2, The following 4 bytes is type of data item in little endian format: 03000000 (tx: 01000000, block: 02000000, filtered block: 03000000, compact block 04000000)

3, the following 32 bytes are hash identifier: 
30eb2540c41025690160a1014c577061596e32e426b712c7ca00000000000000030000001049847939585b0652fba793661c361223446b6fc41089b8be00000000000000

In the payload of getdata message, if we set the type field to value 3, then we are asking the full node to return merkleblock command. Let's code the getdata command, create a new file
getdata.go add the following code:

```go
package bloomfilter

import (
	"math/big"
	"transaction"
)

type Data struct {
	dataTye    []byte
	identifier []byte
}

type GetDataMessage struct {
	command string
	data    []Data
}

func NewGetDataMessage() *GetDataMessage {
	getDataMsg := &GetDataMessage{
		command: "getdata",
		data:    make([]Data, 0),
	}

	return getDataMsg
}

func (g *GetDataMessage) AddData(dataType []byte, identifier []byte) {
	g.data = append(g.data, Data{
		dataTye:    dataType,
		identifier: identifier,
	})
}

func (g *GetDataMessage) Command() string {
	return g.command
}

func (g *GetDataMessage) Serialize() []byte {
	result := make([]byte, 0)
	dataCount := big.NewInt(int64(len(g.data)))
	result = append(result, transaction.EncodeVarint(dataCount)...)
	for _, item := range g.data {
		dataType := new(big.Int)
		dataType.SetBytes(item.dataTye)
		result = append(result, transaction.BigIntToLittleEndian(dataType,
			transaction.LITTLE_ENDIAN_4_BYTES)...)
		result = append(result, transaction.ReverseByteSlice(item.identifier)...)
	}

	return result
}

```
Now we need a block from bitcoin blockchain as our lab rat， go to bitcoin explorer, enter a random number for block height, the explore will search that block for you, I use 199900 as the number:

![截屏2024-07-18 23 26 33](https://github.com/user-attachments/assets/ffafb206-681f-4dad-af9e-e3337f3e1e43)


Click the first item which is the block from the bitcoin mainnet:


![截屏2024-07-18 23 27 07](https://github.com/user-attachments/assets/fc57efd1-85e8-465d-bd0b-3908e180d064)

As you can see there are 197 transactions in this block, copy the block hash and the first transaction hash which is "1df-cc8a" at the right up cornner, save these two strings and we will use them in 
the following code, in simple_node.go we add a new getdata function as following:

```go
func (s *SimpleNode) GetData(conn net.Conn) {
	//prepare bloom filter
	txHash, err := hex.DecodeString("1df77b894e1910628714bb73df59e20fb9114f9dcc051d8c03ca197dd112cc8a")
	if err != nil {
		panic(err)
	}
	bf := bloomfilter.NewBloomFilter(30, 5, 90210)
	//set up bloomfilter ask full node to return any transaction of whichs id
	//map to buckets that have all value 1
	bf.Add(txHash)
	//send filterload
	s.Send(conn, bf.FilterLoadMsg())
	getdata := bloomfilter.NewGetDataMessage()
	receiveMerkleBlock := false
	/*
		ask full node to seach all transaction in the block with the given id,
		put all transaction through the bloom filter we sent by using filterload
		command, then collect all transactions that can pass through the filter,
		put them into merkleblock command
	*/
	blockHash, _ := hex.DecodeString("0000000000000138f016a6fc1666fd667b7d282d65ad14b7f0b16a75a2e90e50")
	getdata.AddData(bloomfilter.FilteredDataType(), blockHash)
	s.Send(conn, getdata)

	for !receiveMerkleBlock {
		//let the peer have a rest
		time.Sleep(2 * time.Second)
		msgs := s.Read(conn)
		for i := 0; i < len(msgs); i++ {
			msg := msgs[i]
			fmt.Printf("receiving command: %s\n", msg.command)
			command := string(bytes.Trim(msg.command, "\x00"))

			if command == "merkleblock" {
				merkleBlock := merkletree.ParseMerkleBlock(msg.payload)
				fmt.Printf("merkleblock received: %s\n", merkleBlock)
				fmt.Printf("merkleblock valid:%v\n", merkleBlock.IsValid())
				receiveMerkleBlock = true
			}

		}
	}
}
```
Then we run the simple node in main.go:
```go
package main

import (
	"networking"
)
//192.168.3.6 is the ip of machine where I run the full node
func main() {
	simpleNode := networking.NewSimpleNode("192.168.3.6", 8333, false)
	simpleNode.Run()
}
```

Running the code above we can get the following result:

```go
eceiving command: merkleblock
merkleblock received: version: 1
previous block: 0000000000000136dbd0fe1bd3d6a7ba4c73ef3ecd46a3e8ac5b1c8c39e4ca84
merkle root: 19dbc2b390c43f9fd34b684a77402f9a8e531dccdc7341be67bf6275be606ef9
bits: 11010001,11011011,10100000,01011000
nonce:fe8539bd
total tx: c5
number of hashes:1
19dbc2b390c43f9fd34b684a77402f9a8e531dccdc7341be67bf6275be606ef9,
flags: 00000000
f96e60be7562bf67be4173dccc1d538e9a2f40774a684bd39f3fc490b3c2db19
merkleblock valid:true
```
As we can see, we receive the merkleblock command, and there is only one hash in the merkle tree, that's because there is only one transaction can pass through the filter and of course it is the
transaction with id "1df77...12cc8a" which is we used to set the filter. In order to include more transactions in the merkleblock we can make some tricks on the filter that is we set one fourth of
the buckets to 1, then we can enable more transactions to pass through the filter:
```go
func (b *BloomFilter) Add(item []byte) {
	for i := 0; i < int(b.funcCount); i++ {
		seed := uint32(uint64(i*BIP37_CONSTANT) + b.tweak)
		h := murmur3.Sum32WithSeed(item, seed)
		idx := h % uint32(len(b.buckets))
		b.buckets[idx] = 1
		fmt.Printf("idx to 1: %d\n", idx)
	}

	//debug set all buckets to 1
	for i := 0; i < len(b.buckets)/4; i++ {
		b.buckets[i] = 1
	}
}
```
Then run the code again we will get the following result:
```go
receiving command: merkleblock
merkleblock received: version: 1
previous block: 0000000000000136dbd0fe1bd3d6a7ba4c73ef3ecd46a3e8ac5b1c8c39e4ca84
merkle root: 19dbc2b390c43f9fd34b684a77402f9a8e531dccdc7341be67bf6275be606ef9
bits: 11010001,11011011,10100000,01011000
nonce:fe8539bd
total tx: c5
number of hashes:15
065f6e836cf36d1bde003a54ab5efa6aea7b37492e77f6e954317bafd64a0382,
8ec46e5599f2f74a61c15267ef683fa3b4793567eeae23df810e481b64b3f1d3,
14ca569146374182e199f1befd39aa5d4ddcfbe3845edb5642da1e97b91ef5b6,
dadbcd9e25b627330cd825eeeb582b739920f645b9ffc38ba2c4e80d57c0a021,
2a2efbe7acd559bb194ce63a0053dabf39db43c5c05f2888883dd51ab8702649,
5fde0ee2a8349044dd37d99956dbe8d1d654c62588649fb8d69bf08245c42bbd,
638012cb174c19fe4384e742acc3ac532408bd5e92310238f3a16fc1e2942059,
99a240b42e42cbd9aaa79cb5f97d265c87e00809d6dc16b68dc848061bba7da5,
ebc98f1fdfc8134ec183bbfdf616fa101a7aa14bd88783bf5f8bdd49bb944a67,
9186b3d8fef79cd5571fd490e1d1eb39dea20d1e5bbea15527df5e07b915f996,
f35fe9feadd9789760b0f1c1cd19a639a4e20f7c439e6e5410fd8bb034bdaf3e,
75ed1ed36c7d89042c96cc2bf092624907f05a87ace47328c60e1eb1f4ef63bf,
1ed1aca69dfaafbe899d0df86bcb58d7313a0a04e2aa64821c94dedbfc7b5f02,
20fe306c0f6e9dd97573a3cc3d66d8f974579e386c6def7444939d1339a6197a,
739c0df831274775dba94820a9ac4dccaf90df8ef5d9307451b23bbadedc41a5,
flags: 11101111110000010101101010100000
f96e60be7562bf67be4173dccc1d538e9a2f40774a684bd39f3fc490b3c2db19
merkleblock valid:true
```
This time we have 15 transactions in the merkleblock, if you set all buckets to 1 then all 197 transactions will include in the merkleblock.

