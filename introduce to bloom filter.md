In previous section, we ask our fullnode peer to return a merbkleblock command that we can verify whether given transactions of interested(the green boxes) are include in a block or not. And most of time we or the client dosen't want the fullnode know which transactions are interested to us, therefore we want to hide our target transactions in a group of transactions(The leafs of the merkle tree). 
Then we need an effective method to transfer info about that group of transactions to fullnode.That's where the data structure and algorithm of bloom filter comes into play.

![image](https://github.com/wycl16514/golang-bitcoin-core-merkle-tree/assets/7506958/8e79a944-601e-4845-be1e-00c4246c6d2f)

As the image shows, we are interesting the green boxes , but we want to hide our intention and we hide these two in a group of transactions that are H(A) to H(P), then we ask full node do you have info about all transaction about H(A) to H(P), 
then the full node will construct the merkleblock command to us, and we can compute the value of the root to verify whether our target transactions that are H(K) and H(N) are put onto the mainet or not.


