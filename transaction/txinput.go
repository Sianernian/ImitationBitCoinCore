package transaction

import (
	"bytes"
)

type TxInput struct{
	Txid [32]byte
	Vout int
	//ScriptSig []byte  //解锁脚本:交易签名，  原始公钥
	// ScriptSig =sig + PubKey
	Sig []byte
	Pubk []byte //
}


func (input *TxInput)CheckPubKeyWithpubKey(pubk []byte)bool{
	return bytes.Compare(input.Pubk,pubk) ==0
}

func NewTxInput(txid [32]byte,vout int,pubk []byte)TxInput{
	input :=TxInput{
		Txid: txid,
		Vout: vout,
		Pubk: pubk,
	}

	return input
}

