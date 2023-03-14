package main

import (
	"log"
	"os"
	"strconv"

	"github.com/shekarramaswamy4/shared-register-abstraction/node"
)

func main() {
	// port
	args := os.Args[1:]
	id, err := strconv.Atoi(args[1])
	if err != nil {
		log.Fatalf("Invalid id: %s", args[1])
	}
	port, err := strconv.Atoi(args[2])
	if err != nil {
		log.Fatalf("Invalid port number: %s", args[2])
	}
	numNodes, err := strconv.Atoi(args[3])
	if err != nil {
		log.Fatalf("Invalid num nodes: %s", args[3])
	}
	numReplicas, err := strconv.Atoi(args[4])
	if err != nil {
		log.Fatalf("Invalid num replicas: %s", args[4])
	}

	n := node.New(id, port, numNodes, numReplicas)

	// OPTIMIZATION: gossip with other nodes to get up to date when boostrapping?
	// should know the possible address space ahead of time

	n.StartHTTP()
}
