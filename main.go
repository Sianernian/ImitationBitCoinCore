package main

import (
	"PublicChain/chain"
	"PublicChain/client"
	"github.com/boltdb/bolt"
)

const BOLTFILE = "pubchain.db"

func main() {

	db, err := bolt.Open(BOLTFILE, 0600, nil)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	blockChain, err := chain.NewBlockChain(db)

	client1 := client.Client{blockChain}
	client1.Run()
}
