package main

import (
	"os"

	"github.com/shekarramaswamy4/shared-register-abstraction/node"
)

func main() {
	// port
	args := os.Args[1:]
	port := args[1]

	n := node.New()

	// TODO: gossip with other nodes to get up to date when boostrapping?
	// should know the possible address space ahead of time

	n.StartHTTP(port)
}
