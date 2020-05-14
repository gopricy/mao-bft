package merkle

import (
	_ "bytes"
	sha256 "crypto/sha256"
	"errors"
	_ "errors"
	pb "github.com/gopricy/mao-bft/pb"
)

// Content defines the object that will get stored in Merkle tree.
type Content interface {
	// This function returns SHA256 hash of the object.
	CalcHash() ([]byte, error)
	// Whether ths content equal another content.
	Equals(content Content) (bool, error)
}

// MerkleTree contains pointer to contents that are stored in it, as well as the tree root.
type MerkleTree struct {
	Root         *Node
	Leaves        []*Node
}

// Node represents the Node that are stored in Merkle tree. A node becomes a leaf when Value is not nil.
type Node struct {
	Hash []byte
	Left *Node
	Right *Node
	Value *Content
}

// Init inits a Merkle tree from contents, it takes in a list of Content, and init the entire tree.
func (tree *MerkleTree) Init(contents []Content) error {
	if len(contents) == 0 {
		return errors.New("content cannot be empty")
	}
	for _, content := range contents {
		hash, err := content.CalcHash()
		if err != nil {
			return errors.New("could not calculate hash from content")
		}
		tree.Leaves = append(tree.Leaves, &Node{
			Hash: hash,
			Value: &content,
		})
	}

	root, err := buildTree(tree.Leaves)
	if err != nil {
		return errors.New("fail to build merkle tree out of ")
	}
	tree.Root = &root
	return nil
}

// buildTree returns a root
func buildTree(nodes []*Node) (Node, error) {
	if len(nodes) == 1 {
		// This is the root, we directly return the root
		return *nodes[0], nil
	}
	var parents []*Node
	for i:=0; i<len(nodes); i+=2 {
		hash := sha256.New()
		// If odds number of nodes, construct the last parent with both left and right as nodes[-1].Hash
		leftNode, rightNode := nodes[i], nodes[i]
		if i+1 < len(nodes) {
			rightNode = nodes[i+1]
		}
		concatHash := append(leftNode.Hash, rightNode.Hash...)
		if _, err := hash.Write(concatHash); err != nil {
			return Node{}, errors.New("cannot hash content: " + string(concatHash))
		}
		parent := Node {
			Hash: hash.Sum(nil),
			Left: leftNode,
			Right: rightNode,
		}
		parents = append(parents, &parent)
	}
	return buildTree(parents)
}

// GetProof returns a MerkleProof for given object. If the supplied object doesn't exist in the tree, return error.
func GetProof(tree *MerkleTree, content Content) (pb.MerkleProof, error) {
	return pb.MerkleProof{}, nil
}