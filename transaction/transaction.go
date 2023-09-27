package transaction

import (
	"PublicChain/utils"
	"PublicChain/wallet"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"time"
)

const REWARD =50


type Transaction struct {
	TxHash  [32]byte //交易的唯一标识
	Inputs  []TxInput
	Outputs []TxOutput
	LockedTime  int64 // 时间戳，为使每笔交易保持唯一性
}

func NewCoinbaseTx(address string)(*Transaction ,error){
	txOutput:=Lock2Address(REWARD,address)
	tx := Transaction{
		Inputs:  []TxInput{},
		Outputs: []TxOutput{txOutput},
		LockedTime:time.Now().Unix(), //
	}

	//序列化
	txBytes, err := utils.GobEncode(tx)
	if err != nil {
		return nil, err
	}

	//交易哈希计算，并赋值给TxHash字段
	tx.TxHash = sha256.Sum256(txBytes)

	return &tx, nil
}

/**
   构建一个新的交易
 */
func NewTransaction(spent []UTXO ,from string , pubk []byte,to string,value float64)(*Transaction ,error){
	//遍历区块 ——>遍历区块中的所有交易
	// 若找到一笔交易，该交易输出A 数额满足需求则return  Txid
	txInputs := make([]TxInput, 0)
	var inputAmount float64
	for _, utxo := range spent {
		inputAmount += utxo.Value
		input :=NewTxInput(utxo.TxId, utxo.Vout,pubk)

		//把构建好的交易输入放入到交易输入容器中
		txInputs = append(txInputs, input)
	}
	//交易输出的容器切片
	//A->B 10 至多会产生两个交易输出
	txOutputs := make([]TxOutput, 0)

	//第一个交易输出：对应转账接收者的输出
	txOutput0:=Lock2Address(value,to)
	txOutputs = append(txOutputs, txOutput0)

	//还有可能产生找零的一个输出：交易发起者给的钱比要转账的钱多
	if inputAmount-value > 0 { //需要找零给转账发起人
		txOutput1 := Lock2Address(inputAmount-value,from)
		txOutputs = append(txOutputs, txOutput1)
	}

	//构建交易
	tx := Transaction{
		Inputs:  txInputs,
		Outputs: txOutputs,
		LockedTime:time.Now().Unix(),
	}
	//序列化
	txBytes, err := utils.GobEncode(tx)
	if err != nil {
		return nil, err
	}
	tx.TxHash = sha256.Sum256(txBytes)

	return &tx, nil
}

/**
   使用私钥对某个交易进行交易的签名
 */
func (tx *Transaction)Sign(private *ecdsa.PrivateKey,utxos []UTXO)(error){

	//交易输入的个数与utxo的个数需要一致
	if len(tx.Inputs) != len(utxos) {
		return errors.New("签名错误")
	}
	txcopy :=CopyTX(*tx)
	for i:=0;i<len(txcopy.Inputs) ;i++{
		input:=txcopy.Inputs[i] //当前遍历到的第几个交易输入
		utxo:=utxos[i]//当前遍历到的第几个utxo
		//scriptPub:=utxo.PubHash //获得当前遍历到的utxo的锁定脚本中的公钥hash
		input.Pubk = utxo.PubHash

		txHash,err :=txcopy.CalculateTxHash()
		if err !=nil{
			return err
		}
		input.Pubk =nil

		r,s,err:=ecdsa.Sign(rand.Reader,private,txHash)
		if err !=nil{
			return err
		}

		sigbytes:=append(r.Bytes(),s.Bytes()...)
		tx.Inputs[i].Sig = sigbytes //赋值的是原tx

	}
	return nil
}

/**
  * 对交易进行验签
 */

func (tx *Transaction) VertifySign(utxos []UTXO) (bool,error){

	if tx.IsCoinbaseTranaction() {
		return true ,nil
	}

	if len(tx.Inputs) !=len(utxos){
		return false,errors.New("验签遇到错误，请检查")
	}
	txCopy :=CopyTX(*tx)

	var result bool
	for index,input:=range txCopy.Inputs{

		// 验签： 公钥，签名，原文 ——> hash
		pubk := input.Pubk    //公钥
		signBytes :=input.Sig // 签名

		// 对交易副本中的每一个input进行还原 ，还原签名之前的状态
		// I. 签名置空
		txCopy.Inputs[index].Sig = nil
		//II. pubk设置为所引用的 utxo的pubkHash
		txCopy.Inputs[index].Pubk =utxos[index].PubHash

		txCopyHash,err :=txCopy.CalculateTxHash()
		if err !=nil{
			return false,err
		}

		// 还原 PublicKey
		//根据[]byte 还原PublicKey
		pub := wallet.GetPublicKeyWithBytes(elliptic.P256(), pubk)

		//调用ecdsa算法库的签名验证方法
		r, s := wallet.RestoreSignature(signBytes)

		result=ecdsa.Verify(&pub,txCopyHash,r,s)

		if !result{//签名验证失败 
			return result,errors.New("签名验证失败")
		}
	}
	return result,nil
}


// 拷贝交易实例
func CopyTX(tx Transaction)(Transaction){
	//制作交易的副本，注意不包含input中的sig字段
	newTx :=Transaction{}
	newTx.TxHash = tx.TxHash

	inputs:=make([]TxInput,0)
	for _,input :=range tx.Inputs{
		txinput:=TxInput{
			Txid: input.Txid,
			Vout: input.Vout,
			Sig:  nil,
			Pubk: input.Pubk,
		}
		inputs =append(inputs,txinput)
	}
	newTx.Inputs =inputs

	//拷贝交易输出
	outputs:=make([]TxOutput,0)
	for _,output:=range tx.Outputs{
		txoutput:=TxOutput{
			Value:   output.Value,
			PubHash: output.PubHash,
		}
		outputs = append(outputs,txoutput)
	}
	newTx.Outputs = outputs

	return newTx
}

/**
  *  判断某个交易是否是Coinbase交易
 */
func (tx *Transaction)IsCoinbaseTranaction() bool {
	return len(tx.Inputs) ==0 && len(tx.Outputs) == 1
}


// 计算交易hash
func (tx *Transaction)CalculateTxHash()([]byte,error){
	txByte,err:=tx.Serialize()
	if err !=nil{
		return nil,err
	}
	return utils.Sha256Hash(txByte) ,nil
}

// 交易的序列化
func (tx *Transaction)Serialize()([]byte,error){
	return utils.GobEncode(tx)
}




