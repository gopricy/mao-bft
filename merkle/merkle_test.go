package merkle
import (
	sha256 "crypto/sha256"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Defines a test content, which contains only 2 bytes array.
type testContent struct {
	x string
}

func (tc *testContent) CalcHash() ([]byte, error) {
	hash := sha256.New()
	if _, err := hash.Write(tc.x[:]); err != nil {
		return nil, errors.New("cannot create hash")
	}
	return hash.Sum(nil), nil
}

func (tc *testContent) Equals(content Content) (bool, error) {
	return len(tc.x) == len(content.(testContent).x), nil
}

func TestBuildTree(t *testing.T) {
	var contents []Content
	contents = append(contents, testContent{x: [2]byte{1, 1}})
	contents = append(contents, testContent{x: [2]byte{2, 2}})

	tree := MerkleTree{}
	tree.Init(contents)
	assert.NotNil(t, tree.Root.Hash)
}