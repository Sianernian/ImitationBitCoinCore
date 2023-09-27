package db

import (
	"PublicChain/chain"
	"github.com/boltdb/bolt"
)

type DBEngine struct {
	DB  *bolt.DB
}

func (engine DBEngine)SaveBlockToDB(block chain.Block){

}


//func (engine DBEngine)GetBlock(hash [32]byte) chain.Block{
//	return nil
//}

