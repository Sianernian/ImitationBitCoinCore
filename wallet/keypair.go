package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"math/big"
)

type KeyPair struct{
	Pri *ecdsa.PrivateKey
	Pub []byte
}

/**
   生成一堆私钥和公钥，返回密钥结构体指针
 */
func NewKeyPair()(*KeyPair,error){
	curve:=elliptic.P256()
	pri,err :=ecdsa.GenerateKey(curve,rand.Reader)
	if err !=nil{
		return nil,err
	}
	pub:=elliptic.Marshal(curve,pri.X,pri.Y)
	keyPair :=&KeyPair{
		Pri: pri,
		Pub: pub,
	}
	return  keyPair,nil
}

/**
 * 使用[]byte类型的数据转换为PublicKey类型公钥
 */
func GetPublicKeyWithBytes(curve elliptic.Curve, data []byte) ecdsa.PublicKey {
	x, y := elliptic.Unmarshal(curve, data)
	return ecdsa.PublicKey{curve, x, y}
}

func RestoreSignature(sign []byte) (r, s *big.Int) {

	rBig := new(big.Int)
	sBig := new(big.Int)
	rBig.SetBytes(sign[:len(sign)/2])
	sBig.SetBytes(sign[len(sign)/2:])

	return rBig, sBig
}

