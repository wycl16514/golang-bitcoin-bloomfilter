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

```
