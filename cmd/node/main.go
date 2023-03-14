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
	port, err := strconv.Atoi(args[1])
	if err != nil {
		log.Fatalf("Invalid port number: %s", args[1])
	}

	n := node.New(port)

	// OPTIMIZATION: gossip with other nodes to get up to date when boostrapping?
	// should know the possible address space ahead of time

	n.StartHTTP()
}
