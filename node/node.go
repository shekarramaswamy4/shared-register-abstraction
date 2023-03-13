package node

import "github.com/google/uuid"

type Node struct {
	ID string
}

func New() *Node {
	return &Node{
		ID: uuid.NewString(),
	}
}

func (n *Node) Read(addr string) (string, error) {
	return "", nil
}

func (n *Node) Write(addr string, val string) error {
	return nil
}

func (n *Node) Confirm(addr string) error {
	return nil
}
