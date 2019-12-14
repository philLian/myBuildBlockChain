package main

import "fmt"

//系列化测试
func TestNewSerialize(){
	//实例化一个区块
	block:=Block{
		2,
		[]byte{},
		[]byte{},
		[]byte{},
		1418755780,
		404454260,
		0,
		[]*Transation{},
		0,
	}

	deBlock:=DeserializeBlock(block.Serialize())
	deBlock.String()
}

//CreatemrekleTreeRoot测试
func TestCreateMerkleTreeRoot(){

	//实例化一个区块
	block:=Block{
		2,
		[]byte{},
		[]byte{},
		[]byte{},
		1418755780,
		404454260,
		0,
		[]*Transation{},
		0,
	}


	//产生第一笔交易
	txin := TXInput{[]byte{},-1,nil,nil}
	txout:=NewTXOutput(subsidy,"first")
	tx :=Transation{nil,[]TXInput{txin},[]TXOutput{*txout}}
	//产生第二笔交易
	txin2 := TXInput{[]byte{},-1,nil,nil}
	txout2:=NewTXOutput(subsidy,"second")
	tx2 :=Transation{nil,[]TXInput{txin2},[]TXOutput{*txout2}}

	//增加2笔交易为Transtions
	var Transtions []*Transation
	Transtions = append(Transtions,&tx,&tx2)

	//将交换构建一个mrekelTree
	block.createMerkelTreeRoot(Transtions)
	fmt.Printf("%x",block.Merkleroot)

}



func TestPow(){
	//实例化一个区块
	block:=&Block{
		2,
		[]byte{},
		[]byte{},
		[]byte{},
		1418755780,
		404454260,
		0,
		[]*Transation{},
		0,
	}

	//实例化一个工作量证明
	pow := NewProofofWork(block)

	nonce,_:=pow.Run()
	block.Nonce =nonce
	fmt.Println("Pow:",pow.Validate() )

}


func TestBoltDB(){
	//添加了创始区块
	blockchain := NewBlockchain("177iKmBCsDZoqJWuBARrZGau7TbbDnXkuB")
	blockchain.MineBlock([]*Transation{})
	blockchain.MineBlock([]*Transation{})
	blockchain.printBlickchain()

}


