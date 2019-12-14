package main

import (
	"bytes"
	"encoding/binary"
	"log"
)

//将一个32b int转化为16进制的字节数组 //以小端模式存
func IntToHex(num int32) []byte {
	//创建一个缓冲区
	buff :=new(bytes.Buffer)
	err :=binary.Write(buff,binary.LittleEndian,num)

	if err!=nil{
		log.Panic(err)
	}

	return buff.Bytes()
}


//将一个32b int转化为16进制的字节数组 //以大端模式存
func IntToHex2(num int32) []byte {
	//创建一个缓冲区
	buff :=new(bytes.Buffer)
	err :=binary.Write(buff,binary.BigEndian,num)

	if err!=nil{
		log.Panic(err)
	}

	return buff.Bytes()
}

//反转2个字节数组
func ReverseBytes(data []byte) {
	for i,j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}

}


//计算两个数的最小值
func Min(a int,b int) int{
	if (a>b){
		return b
	}else {
		return a
	}
}

