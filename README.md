# shared-register-abstraction

Something like https://en.wikipedia.org/wiki/Shared_register

Created a distributed shared register among a configurable amount of nodes. The implementation for the nodes and the client that talks to these nodes are both in this repo. Both nodes and clients are written as web servers.

Nodes and clients are assumed to be non-malicious.

This uses [leaderless replication](https://arpit.substack.com/p/leaderless-replication) to write data. I think it is sequentially consistent, but feel free to prove me wrong :)

## Node
A node has *memory*, which is a mapping from an address to string data.

This implementation assumes a static set of nodes. It is tolerant to network partitions (as long as a quorum is still reachable), but is not designed to handle arbitrary nodes entering and exiting the system.

A node has 4 endpoints: read, write, confirm, and update.

## Client
A client has 2 endpoints: read and write.

## Reads and Writes
Reading data is done by reading from a quorum. Clients fetch data from nodes for a given address and choose the data with the latest confirmed timestamp. Clients then update the out of date nodes.

Writing data is done in two phases, "writing" and "confirming". Both writes and confirms must be acked by a quorum of nodes to declare a write successful.

## Fractional Replication
This implementation also supports fractional replication, where data can only live on a certain subset of the nodes. This parameter is passed into the node constructor.

## Tests
There are unit tests verifying behavior throughout the source code. The most interesting tests are `client_test.go` and `client_fractions_test.go`.

