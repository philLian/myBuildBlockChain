package main

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"
)

var(
	maxnonce int32=math.MaxInt32
)
//构建区块结构体Block
type Block struct{
	Version int32
	PrevBlockHash []byte
	Merkleroot []byte
	Hash []byte
	Time int32
	Bits int32		//难度值
	Nonce int32		//随机数
	Transations []*Transation
	Height int32

}



//对区块数据进行合并系列化
func (b *Block) Serialize() []byte {
	var encoded bytes.Buffer
	//gob:go系列化包
	enc := gob.NewEncoder(&encoded)

	err := enc.Encode(b)
	if err != nil {
		log.Panic(err)
	}
	return encoded.Bytes()
}

//对区块数据进行反系列化
func DeserializeBlock(d []byte)  *Block {
	var block Block

	decode :=gob.NewDecoder(bytes.NewReader(d))

	err :=decode.Decode(&block)
	if err !=nil{
		log.Panic(err)

	}
	return &block
}



//输出区块数据
func (b*Block)String(){
	fmt.Printf("version:%s\n",strconv.FormatInt(int64(b.Version),10))
	fmt.Printf("Prew.BlockHash:%x\n",b.PrevBlockHash)
	fmt.Printf("Merkleroot:%x\n",b.Merkleroot)
	fmt.Printf("Hash:%x\n",b.Hash)
	fmt.Printf("Time:%s\n",strconv.FormatInt(int64(b.Time),10))
	fmt.Printf("Bits:%s\n",strconv.FormatInt(int64(b.Bits),10))
	fmt.Printf("Nonce:%s\n",strconv.FormatInt(int64(b.Nonce),10))
	fmt.Printf("Height:%s\n",strconv.FormatInt(int64(b.Height),10))
}




//bits计算难度值：如18187B74
func CalculateTargetFast(bits []byte) []byte  {
	var result []byte
	//第一个字节：计算指数
	exponent :=bits[:1]
	fmt.Printf("%x\n",exponent)

	//后面三个字节：系数
	coeffient:=bits[1:]
	fmt.Printf("%x\n",coeffient)


	//将字节数组16进制 “18”转化为string
	str:=hex.EncodeToString(exponent)
	fmt.Printf("%s\n",str)

	//将string 转化为10进制 int64 "24"
	exp,_:=strconv.ParseInt(str,16,8)
	fmt.Printf("exp=%d\n",exp)

	//32-exp计算前面要加几个0  计算目标hash
	result = append(bytes.Repeat([]byte{0x00},32-int(exp)),coeffient...)
	//为了保证32个字节，不足补足0
	result = append(result,bytes.Repeat([]byte{0x00},32-len(result))...)


	return result
}


//新建一个merkleTree
func (b *Block) createMerkelTreeRoot(transations []*Transation)  {
	//传递交易hash值
	var tranHash [][]byte
	for _,tx:=range transations{
		tranHash =append(tranHash,tx.Hash())
	}
	//构建merkleTree
	mTree := NewMerkleTree(tranHash)
	b.Merkleroot =mTree.RootNode.Data

}


//新一个非创始区块
func NewBlock(transations []*Transation,prevBlcokHash []byte,height int32) *Block {
	block :=&Block{
		2,
		prevBlcokHash,
		[]byte{},
		[]byte{},
		int32(time.Now().Unix()),
		404454260,
		0,
		transations,
		height,
	}
	//实例化一个工作量证明
	pow :=NewProofofWork(block)
	//挖矿操作
	nonce,hash:=pow.Run()
	block.Nonce=nonce
	block.Hash=hash
	return block

}



//实例化一个创始区块
func NewGensisBlock(transations []*Transation) *Block{
	//实例化一个区块
	block:=&Block{
		2,
		[]byte{},
		[]byte{},
		[]byte{},
		int32(time.Now().Unix()),
		404454260,
		0,
		transations,
		0,
	}
	//实例化一个工作量证明
	pow :=NewProofofWork(block)
	//挖矿操作
	nonce,hash:=pow.Run()
	block.Nonce=nonce
	block.Hash=hash
	//block.String()
	return block

}




