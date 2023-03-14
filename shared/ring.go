package shared

import (
	"hash/fnv"
	"log"
)

// Use the fnv hash function to hash a string to a uint32
// Converts string to int32 determinstically. Not the most random hash function, but
// it's good enough for our purposes.
// "HelloWorld" -> 926844193
func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// includeInShard returns true if the input address should be included in shard.
// We use a modified version of consistent hashing.
// The algorithm is as follows:
// 1) First, break up the address space into numShards shards. Determine which "primary" shard
// the address belongs to by hashing the address and taking the modulus of the number of shards.
// 2) If the primary shard is the shard we're checking, return true.
// 3) If not, the hash belongs to a shard ...
func includeInShard(hash uint32, shard, numShards, numReplicas int) bool {
	// Base case
	if numShards == 1 {
		return true
	}

	primaryShard := int(hash) % numShards
	if shard == primaryShard {
		return true
	} else if numReplicas == 1 {
		return false
	}

	shardStart := primaryShard
	shardEnd := (primaryShard + numReplicas - 1) % numShards

	if shardStart < shardEnd {
		return shardStart < shard && shard <= shardEnd
	} else {
		return shardStart < shard || shard <= shardEnd
	}
}

func HashAndCheckShardInclusion(addr string, shard, numShards, numReplicas int) bool {
	h := hash(addr)
	res := includeInShard(h, shard, numShards, numReplicas)
	log.Printf("Hashing address %s, result is %v, primary shard %d, current shard %d, numReplicas %d, should include %v", addr, h, int(h)%numShards, shard, numReplicas, res)
	return res
}
