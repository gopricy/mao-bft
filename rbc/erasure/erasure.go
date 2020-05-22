package erasure

import (
	"bytes"
	"github.com/gopricy/mao-bft/pb"
	"github.com/gopricy/mao-bft/rbc/merkle"
	"github.com/klauspost/reedsolomon"
	"github.com/pkg/errors"
)

func Split(data []byte, f, t int)([][]byte, error){
	enc, err := reedsolomon.New(t-2*f, 2*f)
	if err != nil{
		return nil, err
	}
	return enc.Split(data)
}

func Reconstruct(payloads []*pb.Payload, f, t int) ([]byte, error){
	shards := make([][]byte, t)
	enc, err := reedsolomon.New(t - 2 * f, 2 * f)
	if err != nil{
		return nil, err
	}
	for _, p := range payloads{
		i := merkle.GetLeafIndex(p.MerkleProof)
		shards[i] = p.Data
	}

	if err := enc.ReconstructData(shards); err != nil{
		return nil, errors.Wrap(err, "Failed to reconstruct the data")
	}
	res := new(bytes.Buffer)
	if err := enc.Join(res, shards, t - 2 * f); err != nil{
		return nil, errors.Wrap(err, "Failed to concat the data")
	}
	return res.Bytes(), nil
}

