package crypto

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
)

func NewKey()( *ecdsa.PrivateKey,error){
	curve :=elliptic.P256()
	return ecdsa.GenerateKey(curve,rand.Reader)
}

func GetPub(pri *ecdsa.PrivateKey )[]byte{
	return bytes.Join([][]byte{
		pri.X.Bytes(),
		pri.Y.Bytes(),
	},[]byte{})
}