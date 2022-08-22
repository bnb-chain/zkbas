# ZkBAS Technology

## ZK Rollup Architecture
![Framework](./assets/Frame_work.png)
- **committer**. Committer executes transactions and produce consecutive blocks.
- **monitor**. Monitor tracks events on BSC, and translates them into **transactions** on zkBAS.
- **witness**. Witness re-executes the transactions within the block and generates witness materials.
- **prover**. Prover generates cryptographic proof based on the witness materials.
- **sender**. The sender rollups the compressed l2 blocks to L1, and submit proof to verify it.
- **api server**. The api server is the access endpoints for most users, it provides rich data, including
  digital assets, blocks, transactions, swap info, gas fees.
- **recovery**. A tool to recover the sparse merkle tree in kv-rocks based on the state world in postgresql.

## Maximum throughput
Pending benchmark...

## Data Availability
ZkBAS publish state data for every transaction processed off-chain to BSC. With this data, it is possible for 
individuals or businesses to reproduce the rollup’s state and validate the chain themselves. BSC makes this data 
available to all participants of the network as calldata.

ZkBAS don't need to publish much transaction data on-chain because validity proofs already verify the authenticity 
of state transitions. Nevertheless, storing data on-chain is still important because it allows permissionless, 
independent verification of the L2 chain's state which in turn allows anyone to submit batches of transactions, 
preventing malicious committer from censoring or freezing the chain.

ZkBAS will provide a default client to replay all state on Layer2 based on these call data.

## Transaction Finality
BSC acts as a settlement layer for ZkBAS: L2 transactions are finalized only if the L1 contract accepts the validity 
proof and execute the txs. This eliminates the risk of malicious operators corrupting the chain 
(e.g., stealing rollup funds) since every transaction must be approved on Mainnet. Also, BSC guarantees that user 
operations cannot be reversed once finalized on L1.

ZkBAS provides relative fast finality speed within 10 minutes.

## Instant confirmation ZkBS
Even though time to finality is about 10 minutes, it does not affect the usability of the network. The state transition
happens immediately once the block been proposed on ZkBAS. The rollup operations are totally transparent to most users, 
users can make further transfers without waiting.

## Censorship resistance
Committer will execute transactions, produce batches. While this ensures efficiency, it increases the risk of censorship
: malicious ZK-rollup committer can censor users by refusing to include their transactions in batches.

As a security measure, ZkBAS allow users to submit transactions directly to the rollup contract on Mainnet if 
they think they are being censored by the operator. This allows users to force an exit from the ZK-rollup to BSC without 
having to rely on the commiter's permission.