package main

import (
	"log"
	"os"
	"strconv"

	"github.com/shekarramaswamy4/shared-register-abstraction/client"
)

func main() {
	// port, numNodes, firstNodePort
	args := os.Args[1:]
	port := args[1]

	numNodes, err := strconv.Atoi(args[2])
	if err != nil {
		log.Fatalf("Invalid number of nodes: %s", args[2])
	}
	if numNodes%2 == 0 {
		log.Fatalf("Number of nodes must be odd")
	}

	// For the purpose of this exercise assume that we can fetch
	// the node ports deterministically from the first node port.
	//
	// We also assume that nodes neither enter or exit the system
	//
	// That is, if there are 3 numNodes, the first node's port will be firstNodePort
	// and the second node's port will be firstNodePort + 1 and so on
	firstNodePort, err := strconv.Atoi(args[3])
	if err != nil {
		log.Fatalf("Invalid first node port: %s", args[3])
	}

	c := client.New(numNodes, firstNodePort)

	c.StartHTTP(port)
}
