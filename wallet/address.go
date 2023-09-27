package wallet

import (
	"PublicChain/utils"
	"bytes"
	"golang.org/x/crypto/ripemd160"
)

func NewAddress(pub []byte)(string , error){

	hashpub:=utils.Sha256Hash(pub)

	Ripemd160:=ripemd160.New()
	Ripemd160.Write(hashpub)
	rip:=Ripemd160.Sum(nil)

	version :=append([]byte{0X00},rip...)

	//fmt.Printf("%x\n",version)
	address := GetAddressWithPubKHash(version)
    return address,nil

}

//据 公钥hash 得到 地址
func GetAddressWithPubKHash(pubkhash []byte)string{
	hash1:=utils.Sha256Hash(pubkhash)
	hash2:=utils.Sha256Hash(hash1)

	code :=hash2[:4]
	addressByte:=append(pubkhash,code...)

	//fmt.Printf("%x\n",addressByte)
	return utils.Encode(addressByte)
}

/**
  用来判断和校验给定的一个字符串是否符合规范
 */
func IsAddressValid(addr string)bool{
	//1. base58反编码
	reverseAdd:=utils.Decode(addr)
	//2. 取反编码后四个字节作为反编码作为校验位
	if len(reverseAdd) < 4 {
		return false
	}
	check := reverseAdd[len(reverseAdd)-4:]

	//3.获取到Versionpub
	versionPub :=reverseAdd[:len(reverseAdd)-4]
	// 4. 双hash
	hash1 :=utils.Sha256Hash(versionPub)
	hash2:=utils.Sha256Hash(hash1)
	code :=hash2[:4]
	//5 比较
	return bytes.Compare(check,code) ==0
}
