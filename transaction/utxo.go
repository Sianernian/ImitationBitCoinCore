package transaction

import (
	"PublicChain/utils"
	"bytes"
)

type UTXO struct {
	TxId [32]byte //表面该可花的钱在哪笔交易
	Vout int      // 表面该可花的钱在该交易的哪个交易输出上
	//Value float64 // 可花费金额的数目 1
	//Owen  string  // 该金额的所有者  2
	TxOutput//用集成TxOUtput方式   等于 1 + 2
}

type SpendReocrdInterface interface {
	GetTxId()  [32]byte
	GetVout()  int
}

func NewUTXO(txid [32]byte,vout int ,out TxOutput)(UTXO){

	utxo:= UTXO{
		TxId:     txid,
		Vout:     vout,
		TxOutput: out,
	}

	return utxo
}

// 某个utxo与传入的utxo进行比较，判断utxo是否花费
func(utxo *UTXO)IsSpent(spend TxInput)bool{
	// utxo.pubkhash ： 公钥hash
	// spend.pubk  ： 原始公钥
	equlTxId :=bytes.Compare(utxo.TxId[:],spend.Txid[:]) == 0

	equalVout := utxo.Vout == spend.Vout
	//把原始公钥 变换 得到 对应公钥hash
	pubk :=spend.Pubk
	pubk256 :=utils.Sha256Hash(pubk)
	ripemd160:= utils.Ripemd160(pubk256)
	versionPubkHash:=append([]byte{0X00},ripemd160...)

	equalLock :=bytes.Compare(versionPubkHash,utxo.PubHash) ==0

	return equlTxId && equalVout && equalLock
}

func(utxo *UTXO) EqualSpendRecord(specified SpendReocrdInterface)bool{
	txid :=specified.GetTxId()
	equalTxID :=bytes.Compare(utxo.TxId[:],txid[:]) == 0
	vout :=specified.GetVout()
	equalVout :=utxo.Vout ==vout

	return equalTxID && equalVout
}