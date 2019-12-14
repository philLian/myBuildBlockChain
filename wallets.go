package main

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const walletFile = "wallet.dat"


//声明钱包的印射表，钱包地址对应私钥和公钥
type Wallets struct{
	Walletsstore map[string]*Wallet

}

//初始化时调用的创建钱包的印射表函数
func NewWallets()(*Wallets,error){
	wallets := Wallets{}
	wallets.Walletsstore = make(map[string]*Wallet)

	err :=wallets.LoadFromFile()

	return &wallets,err

}


//新建钱包并存储于wallets的集合
func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet()
	address:=fmt.Sprintf("%s",wallet.GetAddress())
	ws.Walletsstore[address] =wallet

	return address

}

//获取钱包
func (ws *Wallets) GetWallet(address string) Wallet{

	return *ws.Walletsstore[address]
}



//循环读取生成钱包地址列表
func (ws*Wallets) GetAddress() []string {
	var addresses []string
	for address,_ := range ws.Walletsstore{
		addresses = append(addresses,address)
	}

	return addresses
	
}


//读取文件操作
func (ws *Wallets) LoadFromFile() error {
	if _,err:=os.Stat(walletFile);os.IsNotExist(err){ //判断文件是否存在,如果err不为空，返回True

		return err
	}

	fileContent,err:=ioutil.ReadFile(walletFile)
	if err !=nil{
		log.Panic(err)
	}

	//通过反系列化获取文件内容
	var wallets Wallets
	gob.Register(elliptic.P256()) //调用系列化时的接口
	decoder :=gob.NewDecoder(bytes.NewReader(fileContent)) //实例化反系例化对象
	err =decoder.Decode(&wallets) //执行反系列化操作
	if err !=nil{
		log.Panic(err)
	}

	//将反系列化的结果传递给钱包
	ws.Walletsstore=wallets.Walletsstore

	return nil

}

//将钱包结构体系列化后写入文件walletFile
func (ws *Wallets) SaveToFile(){
	var content bytes.Buffer

	gob.Register(elliptic.P256())	//系列化前调用注册p256曲线 按此接口进行系列化
	encoder:=gob.NewEncoder(&content)
	err :=encoder.Encode(ws)
	if err !=nil{
		log.Panic(err)
	}


	err =ioutil.WriteFile(walletFile,content.Bytes(),0777)  //0777最高权限
	if err !=nil{
		log.Panic(err)
	}
}


