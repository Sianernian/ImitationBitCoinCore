package merkle

import (
	"PublicChain/transaction"
	"math"
)

type MerkleTree struct{
	RootNode  *TreeNode
}

/**
    根据交易生成默克尔树
 */
func GenerateTreeByTransactions(txs []transaction.Transaction)(*MerkleTree,error){
	children,err :=LeafNode(txs)
	if err !=nil{
		return nil,err
	}
	return creatMerkleTree(children),nil
}

/**
  构建一个默克尔树
*/
func creatMerkleTree(leafNode []*TreeNode)(*MerkleTree){
	// 判断叶子的奇偶性
	tree :=MerkleTree{}
	if len(leafNode) %2 ==1{
		leafNode = append(leafNode,leafNode[len(leafNode)-1])
	}

	nowLeveNodes:=leafNode
	//根据叶节点的个数，计算一共需要向上循环生成自己节点
	leveCount :=GetLeveCount(float64(len(leafNode)))
	for level :=0;level<leveCount;level++ {

		leavelNodes := make([]*TreeNode, 0)
		for j := 0; j < len(nowLeveNodes); j += 2 {
			node := CreateTreeNode(nil, leafNode[j], leafNode[j+1])
			leavelNodes = append(leavelNodes, node)
		}

		//当前层数为奇数，复制最后一个节点
		if len(leavelNodes)>1 && len(leavelNodes)%2 == 1 {
			leavelNodes = append(leavelNodes, leavelNodes[len(leavelNodes)-1])
		}
		if len(leavelNodes)==1{
			tree.RootNode = leavelNodes[0]
			break
		}
		nowLeveNodes =leavelNodes
	}
	return &tree

}

/**
   该函数根据叶节点个数，返回共需要循环几次生成子节点
 */
func GetLeveCount(num float64)int {
	var count float64
	count = 0
	for {
	result := math.Pow(2, count)
	if result >=num{
		return int(count)
		}
	}
}