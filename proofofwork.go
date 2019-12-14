package main

import (
	"bytes"
	"crypto/sha256"
	"math/big"
)


//工作量证明结构体
type ProofofWork struct {
	block * Block
	target * big.Int

}

//挖矿难度
const targetBits =16


//实例化一个工作量证明
func NewProofofWork(b *Block) * ProofofWork{
	//目标值初始化为1
	target := big.NewInt(1)
	//目标左移位数
	target.Lsh(target,uint(256-targetBits))
	//fmt.Printf("%x\n",target)
	//初始化工作量证明
	pow :=&ProofofWork{b,target}
	return pow
}


//对区块数据进行合并系列化
func (pow *ProofofWork) prepareDate(nonce int32) []byte{
	data:=bytes.Join(
		[][]byte{
			IntToHex(pow.block.Version),
			pow.block.PrevBlockHash,
			pow.block.Merkleroot,
			pow.block.Hash,
			IntToHex(pow.block.Time),
			IntToHex(pow.block.Bits),
			IntToHex(nonce)},
		[]byte{},


	)
	return data
}

//挖矿操作
func (pow * ProofofWork) Run() (int32,[]byte){

	var nonce int32
	var secodHash [32]byte
	nonce=0
	var currenthash big.Int

	for nonce < maxnonce{
		//系列化
		data:=pow.prepareDate(nonce)
		//double hash
		firstHash :=sha256.Sum256(data)
		secodHash =sha256.Sum256(firstHash[:])
		//fmt.Printf("%x\n",secodHash)

		//将计算后的hash设置成大整数
		currenthash.SetBytes(secodHash[:])
		//如果计算后的值小于目标值
		if currenthash.Cmp(pow.target)==-1{

			break
		}else {
			nonce++
		}
	}

	return nonce,secodHash[:]
}


//测试
func (pow * ProofofWork) Validate() bool{
	var hashInt big.Int
	data:=pow.prepareDate(pow.block.Nonce)
	firstHash :=sha256.Sum256(data)
	secodHash :=sha256.Sum256(firstHash[:])
	hashInt.SetBytes(secodHash[:])
	isValid :=hashInt.Cmp(pow.target) == -1

	return isValid
}

