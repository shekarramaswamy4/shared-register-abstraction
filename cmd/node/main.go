package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/shekarramaswamy4/shared-register-abstraction/node"
)

func main() {
	// port
	args := os.Args[1:]
	port := args[1]

	n := node.New()
	id := n.ID
	nr := &node.NodeResolver{N: n}

	// TODO: gossip with other nodes to get up to date when boostrapping

	fmt.Printf("Running node %s on port %s\n", id, port)
	http.ListenAndServe(fmt.Sprintf(":%s", port), nr)
}
