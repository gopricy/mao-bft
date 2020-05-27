package blockchain

import (
	"github.com/gopricy/mao-bft/pb"
	mao_utils "github.com/gopricy/mao-bft/utils"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestNewLogger(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "*")
	assert.Nil(t, err)

	logger := NewLogger(tmpDir)
	assert.NotNil(t, logger)
	_, err = os.Stat(tmpDir)
	assert.False(t, os.IsNotExist(err))
	assert.Nil(t, err)

	// Clean up.
	os.Remove(tmpDir)
}

func TestLogger_WriteBlockAndReadBack(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "*")
	assert.Nil(t, err)

	logger := NewLogger(tmpDir)
	logger.WriteBlock(pb.Block{CurHash: []byte{0}}, pb.BlockState_BS_STAGED)

	// Force garbage collection.
	logger = nil

	// Read and assert.
	logger = NewLogger(tmpDir)
	dumps, err := logger.ReadAllBlocks()
	assert.Nil(t, err)
	assert.Equal(t, len(dumps), 1)
	assert.Equal(t, dumps[0].State, pb.BlockState_BS_STAGED)
	assert.True(t, mao_utils.IsSameBytes(dumps[0].Block.CurHash, []byte{0}))

	// Clean up.
	os.Remove(tmpDir)
}
