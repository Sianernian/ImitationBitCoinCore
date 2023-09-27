package transaction

import (
	"PublicChain/utils"
	"bytes"
)

type TxOutput struct {
	Value float64
	//ScriptPub  []byte //锁定脚本
	PubHash []byte  //公钥hash
}

/**
   构建一个新的交易输出，锁一定数额
 */
func Lock2Address(value float64,add string)TxOutput{
	reAdd :=utils.Decode(add)
	pubHash:=reAdd[:len(reAdd)-4]
	output :=TxOutput{
		Value:   value,
		PubHash: pubHash,
	}
	return output
}

/**
  该方法用于验证某个交易输出是否是属于某个地址的收入
 */
func(outPut *TxOutput)VertifyOutputWithAddress(add string)(bool){
	reAdd :=utils.Decode(add)
	pubHash:=reAdd[:len(reAdd)-4]
	return bytes.Compare(outPut.PubHash,pubHash) == 0
}