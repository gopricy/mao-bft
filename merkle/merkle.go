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
	// CalcHash returns SHA256 hash of the object.
	CalcHash() ([]byte, error)
	// Equals compare with another another content.
	Equals(content Content) (bool, error)
	// DebugString output this content
	DebugString() string
}

// MerkleTree contains pointer to contents that are stored in it, as well as the tree root.
type MerkleTree struct {
	Root   *Node
	Leaves []*Node
}

// Node represents the Node that are stored in Merkle tree. A node becomes a leaf when Value is not nil.
type Node struct {
	Hash   []byte
	Left   *Node
	Right  *Node
	Parent *Node
	Value  *Content
}

func isSameBytes(left []byte, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	for i, val := range left {
		if right[i] != val {
			return false
		}
	}
	return true
}

func getNodeSibling(node *Node) (*Node, error) {
	if node.Parent == nil {
		return nil, errors.New("parent doesn't have sibling")
	}
	parent := node.Parent
	if isSameBytes(parent.Left.Hash, node.Hash) {
		return parent.Right, nil
	}
	return parent.Right, nil
}

// Init inits a Merkle tree from contents, it takes in a list of Content, and init the entire tree.
func (tree *MerkleTree) Init(contents []Content) error {
	if len(contents) == 0 {
		return errors.New("content cannot be empty")
	}
	for i, content := range contents {
		hash, err := content.CalcHash()
		if err != nil {
			return errors.New("could not calculate hash from content")
		}
		tree.Leaves = append(tree.Leaves, &Node{
			Hash:  hash,
			Value: &contents[i],
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
	for i := 0; i < len(nodes); i += 2 {
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
		parent := Node{
			Hash:  hash.Sum(nil),
			Left:  leftNode,
			Right: rightNode,
		}
		leftNode.Parent, rightNode.Parent = &parent, &parent
		parents = append(parents, &parent)
	}
	return buildTree(parents)
}

// GetProof returns a MerkleProof for given object. If the supplied object doesn't exist in the tree, return error.
func GetProof(tree *MerkleTree, content Content) (pb.MerkleProof, error) {
	for _, leaf := range tree.Leaves {
		if leaf.Value == nil {
			return pb.MerkleProof{}, errors.New("Leaf node cannot contain empty value, node hash: " + string(leaf.Hash))
		}
		// Found same content, construct and return the merkle proof.
		if isEqual, _ := content.Equals(*leaf.Value); isEqual == true {
			return computeMerkleProofFromLeaf(leaf, tree.Root)
		}
	}
	return pb.MerkleProof{}, errors.New("does not find the content in tree: " + content.DebugString())
}

func computeMerkleProofFromLeaf(node *Node, root *Node) (pb.MerkleProof, error) {
	if root.Parent != nil {
		return pb.MerkleProof{}, errors.New("root is invalid, it contains parent")
	}
	proof := pb.MerkleProof{Root: root.Hash}
	curNode := node
	for curNode.Parent != nil {
		sibling, err := getNodeSibling(curNode)
		if err != nil {
			return pb.MerkleProof{}, err
		}
		// Add proof pair into ProofPairs, in a bottom up way.
		proof.ProofPairs = append(proof.ProofPairs, &pb.ProofPair{
			Primary:   curNode.Hash,
			Secondary: sibling.Hash,
		})
		curNode = curNode.Parent
	}
	// validate that now curNode should be root
	if !isSameBytes(curNode.Hash, root.Hash) {
		return pb.MerkleProof{}, errors.New("early abort before reaching parent, curNode hash is: " + string(curNode.Hash))
	}
	return proof, nil
}

// verifyHashToParent verifies that hash(left + right) == parent.
func verifyHashToParent(left []byte, right []byte, parent []byte) bool {
	hash := sha256.New()
	if _, err := hash.Write(append(left, right...)); err != nil {
		return false
	}
	return isSameBytes(hash.Sum(nil), parent)
}

// VerifyProof verifies a MerkleProof, it takes in a data list, and verify all the way to the end.
func VerifyProof(proof pb.MerkleProof, content Content) bool {
	// If proof just contain root, it should be a single node tree.
	if proof.ProofPairs == nil {
		contentHash, err := content.CalcHash()
		return err == nil && isSameBytes(contentHash, proof.Root)
	}

	// Verify all the way to root hash.
	for i, proofPair := range proof.ProofPairs {
		if i+1 < len(proof.ProofPairs) {
			parentPair := proof.ProofPairs[i+1]
			if !verifyHashToParent(proofPair.Primary, proofPair.Secondary, parentPair.Primary) {
				return false
			}
		} else {
			if !verifyHashToParent(proofPair.Primary, proofPair.Secondary, proof.Root) {
				return false
			}
		}
	}

	// Verify content hashes to first primary.
	contentHash, err := content.CalcHash()
	if err != nil || !isSameBytes(contentHash, proof.ProofPairs[0].Primary) {
		return false
	}
	return true
}
