package merkle

import (
	sha256 "crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Defines a test content, which contains a string as value.
type testContent struct {
	x string
}

func (tc *testContent) CalcHash() ([]byte, error) {
	hash := sha256.New()
	if _, err := hash.Write([]byte(tc.x)); err != nil {
		return nil, errors.New("cannot create hash")
	}
	return hash.Sum(nil), nil
}

func (tc *testContent) Equals(content Content) (bool, error) {
	return tc.x == content.(*testContent).x, nil
}

func (tc *testContent) String() string {
	return tc.x
}

func TestBuildTreeSingleNode(t *testing.T) {
	var contents []Content
	contents = append(
		contents,
		&testContent{x: "a"})
	tree := MerkleTree{}
	tree.Init(contents)
	assert.NotNil(t, tree.Root.Hash)
	hashHexString := hex.EncodeToString(tree.Root.Hash)
	assert.Equal(t, "ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb", hashHexString)
}

func TestBuildTreeEvenNodes(t *testing.T) {
	var contents []Content
	contents = append(
		contents,
		&testContent{x: "a"}, &testContent{x: "b"})
	tree := MerkleTree{}
	tree.Init(contents)
	assert.NotNil(t, tree.Root.Hash)
	hashHexString := hex.EncodeToString(tree.Root.Hash)
	assert.Equal(t, "e5a01fee14e0ed5c48714f22180f25ad8365b53f9779f79dc4a3d7e93963f94a", hashHexString)
	assert.Equal(t, len(tree.Leaves), 2)
	assert.Equal(t, (*tree.Leaves[0].Value).String(), "a")
}

func TestBuildTreeOddNodes(t *testing.T) {
	var contents []Content
	contents = append(
		contents,
		&testContent{x: "a"}, &testContent{x: "b"}, &testContent{x: "c"})
	tree := MerkleTree{}
	tree.Init(contents)
	assert.NotNil(t, tree.Root.Hash)
	hashHexString := hex.EncodeToString(tree.Root.Hash)
	assert.Equal(t, "d31a37ef6ac14a2db1470c4316beb5592e6afd4465022339adafda76a18ffabe", hashHexString)
}

func TestGetMerkleProof(t *testing.T) {
	var contents []Content
	contents = append(
		contents,
		&testContent{x: "a"}, &testContent{x: "b"}, &testContent{x: "c"})
	tree := MerkleTree{}
	tree.Init(contents)
	merkleProof, err := GetProof(&tree, &testContent{x: "a"})
	assert.Nil(t, err)
	assert.Equal(t,
		"d31a37ef6ac14a2db1470c4316beb5592e6afd4465022339adafda76a18ffabe",
		hex.EncodeToString(merkleProof.Root))
	assert.Equal(t,
		"ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb",
		hex.EncodeToString(merkleProof.ProofPairs[0].Primary))
}

func TestGetMerkleProofSingleNode(t *testing.T) {
	var contents []Content
	contents = append(
		contents,
		&testContent{x: "a"})
	tree := MerkleTree{}
	tree.Init(contents)
	merkleProof, err := GetProof(&tree, &testContent{x: "a"})
	assert.Nil(t, err)
	assert.Nil(t, merkleProof.ProofPairs)
	assert.Equal(t,
		"ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb",
		hex.EncodeToString(merkleProof.Root))
}

func TestVerifyProofShouldReturnTrueForGetProof(t *testing.T) {
	var contents []Content
	contents = append(
		contents,
		&testContent{x: "a"}, &testContent{x: "b"}, &testContent{x: "c"})
	tree := MerkleTree{}
	tree.Init(contents)
	merkleProof, _ := GetProof(&tree, &testContent{x: "a"})
	assert.Equal(t, true, VerifyProof(merkleProof, &testContent{x: "a"}))

	// Test negative.
	assert.Equal(t, false, VerifyProof(merkleProof, &testContent{x: "b"}))
}

func TestVerifyProofSingleNode(t *testing.T) {
	var contents []Content
	contents = append(
		contents,
		&testContent{x: "a"})
	tree := MerkleTree{}
	tree.Init(contents)
	merkleProof, _ := GetProof(&tree, &testContent{x: "a"})
	assert.Equal(t, true, VerifyProof(merkleProof, &testContent{x: "a"}))

	// Test negative.
	assert.Equal(t, false, VerifyProof(merkleProof, &testContent{x: "b"}))
}
