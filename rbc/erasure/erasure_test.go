package erasure_test

import (
	"math/rand"
	"testing"

	"github.com/gopricy/mao-bft/rbc/erasure"
	"github.com/stretchr/testify/assert"
)

var testbytes []byte

func init() {
	testbytes = testData()
}

func testData() []byte {
	res := make([]byte, rand.Intn(1000))
	rand.Read(res)
	return res
}

func TestSplit(t *testing.T) {
	f := rand.Intn(5)
	n := 3*f + 1
	shards, err := erasure.Split(testbytes, f, n)
	assert.Nil(t, err)
	assert.Equal(t, n, len(shards))
	data := []byte{}
	for i := 0; i < n-2*f; i++ {
		data = append(data, shards[i]...)
	}
	for {
		if data[len(data)-1] == byte(0) {
			data = data[:len(data)-1]
		} else {
			break
		}
	}
	assert.Equal(t, testbytes, data)
}

func TestReconstruct(t *testing.T) {
	f := rand.Intn(5)
	n := 3*f + 1
	shards, err := erasure.Split(testbytes, f, n)
	assert.Nil(t, err)
	assert.Equal(t, n, len(shards))
	data := []byte{}
	// set 2 * f shards to nil
	for i := 0; i < 2*f; {
		r := rand.Intn(len(shards))
		if shards[r] != nil {
			shards[r] = nil
			i += 1
		}
	}
	data, err = erasure.ReconstructBytes(shards, f)
	assert.Nil(t, err)
	assert.Equal(t, testbytes, data)
}
