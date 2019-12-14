package main

import "crypto/sha256"




//MerkleTree根结构体
type MerkleTree struct {
	RootNode *MerkleNode
}


//MerkleTree子结点结构体
type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

//MerkleTree对左右结点数据进行合并及hash运算
func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	mnode := MerkleNode{}
	//如果左右为空就是根结点
	if left == nil && right == nil {
		mnode.Data = data
	} else {
		//否则进行左右合并并进行2次hash256
		prevhashes := append(left.Data, right.Data...)
		firsthash := sha256.Sum256(prevhashes)
		hash := sha256.Sum256(firsthash[:])
		mnode.Data = hash[:]

	}
	//左右节点的MerkleTree不变
	mnode.Left = left
	mnode.Right = right

	//返回左右合并后的MerkleTree根结点
	return &mnode

}

//构建merkleTree
func NewMerkleTree(data [][]byte) *MerkleTree {

	//声明一个MerkleNode结构体变量
	var nodes []MerkleNode

	//循环data 取每个节点的hash值append至nodes
	for _, datum := range data {
		node := NewMerkleNode(nil, nil, datum)
		nodes = append(nodes, *node)
	}

	//判断传入的区块hash有几个，然后循环至nSize等于1时结束（1也就是到了根结点）
	j:=0
	for nSize :=len(data);nSize>1;nSize=(nSize+1)/2{
		//对每个传入的切片进行22循环
		for i:=0;i<nSize;i+=2{
			//求i2是不是最后当次循环的最后一个，如果是最后一个就返回最后一个
			i2:=Min(i+1,nSize-1)
			node :=NewMerkleNode(&nodes[j+i],&nodes[j+i2],nil)
			nodes = append(nodes,*node)
		}
		//第一层子结点结束改变循环开始位置j的值，开始新的循环，直至没有nSize>1结束（表示已到了根结点）
		j+=nSize
	}
	//取最后的结点就是MerkleTree
	mTree := MerkleTree{&(nodes[len(nodes)-1])}
	return &mTree
}
