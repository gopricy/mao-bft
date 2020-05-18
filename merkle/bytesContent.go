package merkle

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

type BytesContent []byte

func (bc BytesContent) CalcHash() ([]byte, error) {
	hash := sha256.New()
	if _, err := hash.Write(bc); err != nil {
		return nil, errors.New("cannot create hash")
	}
	return hash.Sum(nil), nil
}

func (bc BytesContent) Equals(content Content) (bool, error) {
	t, ok := content.(BytesContent)
	if !ok{
		// different underlying type
		return false, nil
	}

	if len(bc) != len(t){
		return false, nil
	}
	for i := range(bc){
		if bc[i] != t[i]{
			return false, nil
		}
	}
	return true, nil
}

func (bc BytesContent) String() string {
	return hex.EncodeToString(bc)
}

