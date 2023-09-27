package chain

import (
	"PublicChain/consensus"
	"PublicChain/merkle"
	"PublicChain/transaction"
	"bytes"
	"encoding/gob"
	"time"
)

type Block struct{
	Height  int64   //4
	Version int64    //4
	PreHash [32]byte   //32
	Hash    [32]byte
	// 默克尔根
	MerkleRoot []byte    //32
	Timestamp int64  // 4
	//Difficulty int64
	Nonce int64  // 4
	//Data  []byte //区块体
	Txs []transaction.Transaction

}


/**
   区块的序列化，序列化为 []byte数据类型
 */
func(block *Block)Serialize()([]byte ,error){
	buffer :=new(bytes.Buffer)
	encoder :=gob.NewEncoder(buffer)
	err :=encoder.Encode(&block)
	return buffer.Bytes() , err

}

/**
   反序列化操作，传入[]byte，返回Block结构体
 */
func UnSerialize(data []byte)(Block ,error){
	var block Block
	encoder :=gob.NewDecoder(bytes.NewReader(data))
	err :=encoder.Decode(&block)
	return block , err
}


func CreateBlock(height int64 ,prevHash [32]byte ,txs []transaction.Transaction)(*Block,error){
	block :=Block{}
	block.Height =height + 1
	block.PreHash = prevHash
	block.Version = 0X00
	block.Timestamp = time.Now().Unix()
	block.Txs = txs

	//调用生成merkle树
	tree,err:=merkle.GenerateTreeByTransactions(txs)
	if err !=nil{
		return nil,err
	}
	block.MerkleRoot = tree.RootNode.Value


	proof := consensus.NewProofWork(block)
	hash, nonce := proof.SearchNonce()
	block.Nonce = nonce

	block.Hash = hash

	return &block,nil
}


func CreateGenesisBlock(txs []transaction.Transaction)Block{
	genesis :=Block{}
	genesis.Height =0
	genesis.PreHash =[32]byte{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0}
	genesis.Version = 0X01
	genesis.Timestamp = time.Now().Unix()
	genesis.Txs =txs

	proof := consensus.NewProofWork(genesis)
	hash, nonce := proof.SearchNonce()
	genesis.Hash = hash
	genesis.Nonce = nonce

	return genesis
}

/**
 * 该方法是实现BlockInterface的GetHeight方法
 */
func (block Block) GetHeight() int64 {
	return block.Height
}

/**
 * 该方法是实现BlockInterface的GetVersion方法
 */
func (block Block) GetVersion() int64 {
	return block.Version
}

func (block Block) GetTimeStamp() int64 {
	return block.Timestamp
}

func (block Block) GetPreHash() [32]byte {
	return block.PreHash
}

func (block Block) GetTxs()  []transaction.Transaction {
	return block.Txs
}

func(block Block)GetMerkleRoot()[]byte{
	return block.MerkleRoot
}