package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
)

const nodeversion = 0x00

var nodeAddress string
var blockInTransit [][]byte
const commonLength = 12

type Version struct {
	Version    int
	BaseHeight int32
	AddrFrom   string
}

func (ver *Version) String()  {
	fmt.Printf("Version:%d\n",ver.Version)
	fmt.Printf("BaseHeight:%d\n",ver.BaseHeight)
	fmt.Printf("AddrFrom:%s\n",ver.AddrFrom)

}

var knownNodes = []string{"localhost:3000"}

//开启服务
func StartServer(nodeID, minerAddress string,bc *Blockchain) {
	//本地端口
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	//用tcp端口监听

	ln, err := net.Listen("tcp", nodeAddress)
	fmt.Println("a1")
	//断开后关闭连接
	defer ln.Close()
	//bc := NewBlockchain("177iKmBCsDZoqJWuBARrZGau7TbbDnXkuB")

	//如果本地节点不等于已知结节发送区块链对象
	if nodeAddress != knownNodes[0] {
		sendVersion(knownNodes[0], bc)
	}
	fmt.Println("a2")

	//不断监听
	for {
		fmt.Println("a3")

		conn, err2 := ln.Accept() //一直等待接收消息
		fmt.Println("a4")

		if err2 != nil {
			log.Panic(err)
		}

		go handleConnction(conn, bc) //开启协程处理链接
	}
}

//对收到节点消息进行处理
func handleConnction(conn net.Conn, bc *Blockchain) {
	request, err := ioutil.ReadAll(conn)
	if err != nil{
		log.Panic(err)
	}
	//解析指令
	command := bytesTocommand(request[:commonLength])
	println(command)
	switch command {
	case "version":
		fmt.Println("获取version指令")
		handleVersion(request, bc)
	case "getblocks":
		handleGetBlock(request,bc)
	case "inv":
		handleInv(request,bc)
	case "getdata":
		handleGetData(request,bc)
	case "block":
		handleBlock(request,bc)
	}
}

func handleBlock(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload blocksend
	buff.Write(request[commonLength:])
	dec :=gob.NewDecoder(&buff)
	err :=dec.Decode(&payload)
	if err !=nil{
		log.Panic(err)
	}
	blockdata:=payload.Block
	block :=DeserializeBlock(blockdata)
	bc.AddBlock(block)
	fmt.Printf("Recieve a new Block")

	if len(blockInTransit) >0{
		blockHash := blockInTransit[0]
		sendGetDate(payload.AddrFrom,"block",blockHash)
		blockInTransit = blockInTransit[1:]
	}else{
		set:=UTXOSet{bc}
		set.Reindex()
	}
}

func handleGetData(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload getdata
	buff.Write(request[commonLength:])
	dec:=gob.NewDecoder(&buff)
	err:=dec.Decode(&payload)
	if err !=nil{
		log.Panic(err)
	}
	if payload.Type == "block"{
		block,err:=bc.GetBlock([]byte(payload.ID))
		if err !=nil{
			log.Panic(err)
		}
		sendBlock(payload.AddrFrom,&block)
	}
}

type blocksend struct {
	AddrFrom string
	Block []byte

}

func sendBlock(addr string, block *Block) {
	data:=blocksend{nodeAddress,block.Serialize()}
	payload := gobEncode(data)
	request:=append(commandToBytes("block"),payload...)
	sendData(addr,request)
}

func handleInv(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload inv
	buff.Write(request[commonLength:])
	dec := gob.NewDecoder(&buff)
	err:=dec.Decode(&payload)
	if err!=nil{
		log.Println(err)
	}

	fmt.Printf("Recieve inventory %d,%s",len(payload.Items),payload.Type)

	if payload.Type == "block"{
		blockInTransit=payload.Items
		blockHash := payload.Items[0]
		sendGetDate(payload.AddrFrom,"block",blockHash)
		newInTransit:=[][]byte{}
		for _,b:=range blockInTransit{
			if bytes.Compare(b,blockHash)!=0{
				newInTransit = append(newInTransit,b)
			}
		}
		blockInTransit =newInTransit
	}
}

type getdata struct {
	AddrFrom string
	Type string
	ID []byte
}

func sendGetDate(addr string, kind string, id []byte) {
	payload:=gobEncode(getdata{nodeAddress,kind,id})
	request := append(commandToBytes("getdata"),payload...)
	sendData(addr,request)
}



func handleGetBlock(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload getblocks
	buff.Write(request[commonLength:])
	dec:=gob.NewDecoder(&buff)
	err:=dec.Decode(&payload)
	if err !=nil{
		log.Println(err)
	}
	block :=bc.GetBlockhash()
	sendInv(payload.Addrfrom,"block",block)

}

type inv struct {
	AddrFrom string
	Type string
	Items [][]byte
}

//传递数据至其它节点
func sendInv(addr string, kind string, items [][]byte) {
	inventory:=inv{nodeAddress,kind,items}
	payload := gobEncode(inventory)
	request := append(commandToBytes("inv"),payload...)
	sendData(addr,request)

}



//对收到的“version”指令处理
func handleVersion(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload Version
	buff.Write(request[commonLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)

	if err != nil {
		log.Panic(err)
	}

	payload.String()

	myBestHeight :=bc.GetBestHeight()
	foreignerBestHeight :=payload.BaseHeight

	if myBestHeight < foreignerBestHeight{
		sendGetBlock(payload.AddrFrom)
	}else{
		sendVersion(payload.AddrFrom,bc)
	}

	//对收到的地址如果没有在我我已知列表中，就添加到知列表中
	if !nodeIsKnow(payload.AddrFrom){
		knownNodes = append(knownNodes,payload.AddrFrom)
	}

}

type getblocks struct {
	Addrfrom string
}

//获取最新区块节点
func sendGetBlock(address string) {
	payload:=gobEncode(getblocks{nodeAddress})
	request:= append(commandToBytes("getblocks"),payload...)
	sendData(address,request)
}


//判断节点是否在我的列表当中
func nodeIsKnow(addr string) bool {
	for _,node :=range knownNodes{
		if node == addr{
			return true
		}
	}
	return false
}




//发送区块链对象
func sendVersion(addr string, bc *Blockchain) {
	bestHeight := bc.GetBestHeight()

	//payload系列化构建了要发送的数据
	payload := gobEncode(Version{nodeversion, bestHeight, nodeAddress})

	//合并发送内容指令
	request := append(commandToBytes("version"), payload...)
	sendData(addr, request)
}

//发关消息
func sendData(addr string, data []byte) {
	con, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Printf("%s is not available", addr)

		//更新节点
		var updateNodes []string
		for _, node := range knownNodes {
			if node != addr {
				updateNodes = append(updateNodes, node)
			}
		}
		knownNodes = updateNodes
	}

	//断开后关闭连接
	defer con.Close()
	//发送资源
	_, err = io.Copy(con, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

//转化一个固定长度的字节数组
func commandToBytes(command string) []byte {
	var bytes [commonLength]byte
	for i, c := range command {
		bytes[i] = byte(c)
	}
	return bytes[:]
}

//反转字节数组为string
func bytesTocommand(bytes []byte) string {
	var command []byte
	for _, b := range bytes {
		if b != 0x00 {
			command = append(command,b)
		}
	}
	return fmt.Sprintf("%s", command)
}

//系列化接口
func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}
