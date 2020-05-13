package merkel

import (
	_ "bytes"
	_ "crypto/sha256"
	_ "errors"
	_ "hash"
)

// This defines the object that will get stored in Merkle tree.
type Content interface {
	// This function returns SHA256 hash of the object.
	CalcHash() ([]byte, error)
	// Whether ths content equal another content.
	Equals(content Content) (bool, error)
}

// The Merkle tree object. It contains pointer to contents that are stored in it, as well as the tree root.
type MerkleTree struct {
	Root         *Node
	Leafs        []*Node
}

// This represents the Node that are stored in Merkle tree. A node becomes a leaf when Value is not nil.
type Node struct {
	Hash []byte
	Left *Node
	Right *Node
	Value Content
}

// This function inits a Merkle tree, it takes in a list of Content, and init the entire tree.
func (tree *MerkleTree) Init(contents []Content) error {
	return nil
}

// This function returns a proof for
func GetProof(tree *MerkleTree, content Content) (MerkleProof, error) {
	return nil, nil
}