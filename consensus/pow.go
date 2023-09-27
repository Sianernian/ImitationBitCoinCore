package consensus

import (
	"PublicChain/utils"
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
)


const DIFFCULT = 20

/**
 * 工作量证明
 */
type ProofWork struct {
	Block  BlockInterface
	Target *big.Int
}

/**
 * 实现共识机制接口的方法
 */
func (work ProofWork) SearchNonce() ([32]byte, int64) {

	//1 给定一个non值，计算带有non的区块哈希
	var nonce int64
	nonce = 0
	hashBig :=new(big.Int)

	for {
		hash := CalculateBlockHash(work.Block, nonce)
		//2 系统给定的值
		target := work.Target
		//3 拿1和2比较
		hashBig =hashBig.SetBytes(hash[:])
		result :=hashBig.Cmp(target)
		//result := bytes.Compare(hash[:], target.Bytes())
		//4 判断结果，区块哈希<给定值，返回non;
		if result == -1 {
			return hash, nonce
		}
		//否则，non自增
		nonce++
	}
}

/**
 * 根据当前的区块和当前的non值，计算区块的哈希值
 */
func CalculateBlockHash(block BlockInterface, nonce int64) [32]byte {
	heightByte, _ := utils.IntToByte(block.GetHeight())
	versionByte, _ := utils.IntToByte(block.GetVersion())
	timeByte, _ := utils.IntToByte(block.GetTimeStamp())
	nonceByte, _ := utils.IntToByte(nonce)
	preHash := block.GetPreHash()
	txs :=block.GetTxs()
	txsBytes :=make([]byte,0)
	for _,tx:=range txs{
		tyBytes ,err :=utils.GobEncode(tx)
		if err !=nil{
			fmt.Println(err.Error())
		}
		txsBytes =append(txsBytes,tyBytes...)
	}

	bk := bytes.Join([][]byte{heightByte,
		versionByte,
		preHash[:],
		timeByte,
		nonceByte,
		txsBytes,
		},
		[]byte{})
	return sha256.Sum256(bk)
}
