package sign

import (
	inner "golang.org/x/crypto/nacl/sign"
)

type PrivateKey *[64]byte
type PublicKey *[32]byte

func GenerateKey() (PublicKey, PrivateKey){
	pub, prv, err := inner.GenerateKey(nil)
	if err != nil{
		panic(err)
	}
	return pub, prv
}

func Sign(privateKey PrivateKey, message []byte) []byte{
	return inner.Sign(nil, message, privateKey)
}

func Verify(publicKey PublicKey, message []byte)(decoded []byte, verified bool){
	return inner.Open(nil, message, publicKey)
}
