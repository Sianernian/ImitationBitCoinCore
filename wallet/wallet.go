package wallet

import (
	"PublicChain/utils"
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"github.com/boltdb/bolt"
)

const KEYSTORE = "keystore"
const COINBASE = "coinbase"
// key
const ADDRESS  = "address_keypair"

type Wallet struct {
	Address  map[string]*KeyPair
	DB       *bolt.DB

}

func(wallet *Wallet)CreateNewAddress()(string,error){

	keypair,err:=NewKeyPair()
	if err !=nil{
		return "",err
	}

	address,err :=NewAddress(keypair.Pub)
	if err !=nil{
		return "",err
	}
	wallet.Address[address] = keypair
	err = wallet.SaveMenToDB()

	return address , err
}

/**
   把新生成的地址和对应的KeyPair写入DB中
 */

func (wallet *Wallet)SaveMenToDB()error{
	var err error

	wallet.DB.Update(func(tx *bolt.Tx) error {
		bucket:=tx.Bucket([]byte(KEYSTORE))

		if bucket ==nil{
			bucket,err =tx.CreateBucket([]byte(KEYSTORE))
			if err !=nil{
				return err
			}
		}
		// 把内存中的地址和对应的密钥信息存入db中
		//for key,value:=range wallet.Address  {
		//	keyBytes:=bucket.Get([]byte(key))
		//	if len(keyBytes) == 0{
		//		keyPairBytes,err:=utils.GobEncode(value)
		//		if err !=nil{
		//			return err
		//		}
		//		bucket.Put([]byte(key),keyPairBytes)
		//	}
		//}
		gob.Register(elliptic.P256())

		addAndKeyPairbytes,err :=utils.GobEncode(wallet.Address)
		if err !=nil{
			return err
		}
		bucket.Put([]byte(ADDRESS),addAndKeyPairbytes)
		return nil
	})
	return  err
}

/**
  从db文件中加载数据，构建wallet结构体实例
 */
func LoadWalletFromDB(db *bolt.DB) (Wallet,error){
	var wallet Wallet
	var err error
	//var adds map[string]*KeyPair
	adds :=make(map[string]*KeyPair)

	db.View(func(tx *bolt.Tx) error {
		bucket:=tx.Bucket([]byte(KEYSTORE))
		if bucket ==nil{
			return nil
		}
		addAndKeyPairBytes:=bucket.Get([]byte(ADDRESS))

		if len(addAndKeyPairBytes) ==0{
			return nil
		}
		//反序列化map
		gob.Register(elliptic.P256())
		decoder :=gob.NewDecoder(bytes.NewReader(addAndKeyPairBytes))
		err = decoder.Decode(&adds)
		if err !=nil{
			return err
		}
		return nil
	})
	//实例化结构体，并赋值
	wallet =Wallet{
		Address: adds,
		DB:      db,
	}
	return wallet,err
}

// 根据地址取出 公钥
func (wallet *Wallet)GetKeyPairByAddress(address string)(*KeyPair){
	return wallet.Address[address]
}

func (wallet *Wallet)SetCoinbase(address string) (error){
	var err error
	wallet.DB.Update(func(tx *bolt.Tx) error {
		bucket:=tx.Bucket([]byte(KEYSTORE))
		if bucket ==nil{
			bucket,err = tx.CreateBucket([]byte(KEYSTORE))
			if err !=nil{
				return err
			}
		}
		bucket.Put([]byte(COINBASE),[]byte(address))
		return nil
	})
	return err
}


/// 封装 钱包的 获取当前矿工地址功能
func(wallt *Wallet)GetCoinbase()string{

	var coinbase []byte
	wallt.DB.Update(func(tx *bolt.Tx) error {
		bucket :=tx.Bucket([]byte(KEYSTORE))
		if  bucket ==nil{
			return nil
		}
		coinbase=bucket.Get([]byte(COINBASE))
		return nil
	})
	return string(coinbase)
}

// 根据 公钥hash 得到 地址
func(wallet *Wallet)GetAddressByPubKHash(data []byte)string{
	return GetAddressWithPubKHash(data)
}



