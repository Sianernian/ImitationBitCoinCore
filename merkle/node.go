package merkle

import (
	"PublicChain/transaction"
	"PublicChain/utils"
	"errors"
)

/*
*

	定义默克尔树节点
*/
type TreeNode struct {
	Value     []byte
	LeftNode  *TreeNode
	RightNode *TreeNode
}

func LeafNode(txs []transaction.Transaction) ([]*TreeNode, error) {
	if len(txs) <= 0 {
		return nil, nil
	}
	// 交易个数为奇数，复制最后有一个交易
	if len(txs)%2 == 1 {
		txs = append(txs, txs[len(txs)-1])
	}
	//遍历交易切片，每一个交易构建一个 叶节点
	leafNodes := make([]*TreeNode, 0)
	for _, tx := range txs {
		txSerialize, err := tx.Serialize()
		if err != nil {
			return nil, errors.New("交易为空")
		}
		txHash := utils.Sha256Hash(txSerialize)
		leafNode := CreateTreeNode(txHash, nil, nil)
		leafNodes = append(leafNodes, leafNode)
	}
	return leafNodes, nil
}

/*
*

	构建一个新的节点 并返回
*/
func CreateTreeNode(hash []byte, right *TreeNode, left *TreeNode) *TreeNode {
	node := TreeNode{}
	if right == nil && left == nil {
		node.Value = hash
		node.LeftNode = nil
		node.RightNode = nil

	} else { //当前要生成的节点  非 叶节点
		data := make([]byte, 0)
		data = append(data, right.Value[:]...)
		data = append(data, left.Value[:]...)

		hash1 := utils.Sha256Hash(data)
		hash2 := utils.Sha256Hash(hash1)

		node.Value = hash2
		//描述左右节点
		node.LeftNode = left
		node.RightNode = right
	}
	return &node
}
