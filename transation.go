package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"log"
	"math/big"
)

const subsidy = 100

type Transation struct{
	ID []byte
	Vin []TXInput
	Vout []TXOutput
}

//输入
type TXInput struct{
	TXid []byte
	Voutindex int
	Signature []byte

	Pubkey []byte	//公钥
}

//输出
type TXOutput struct{
	Value int
	PubkeyHash []byte	//公钥的hash
}

//存储所有输出的切片
type TXOutputs struct{
	Outputs []TXOutput
}

//系例化方法
func (outs TXOutputs) Serializa()[]byte  {
	var buff bytes.Buffer
	enc:=gob.NewEncoder(&buff)
	err:=enc.Encode(outs)
	if err!=nil{
		log.Panic(err)

	}
	return buff.Bytes()
}

//反系例化操作
func DeserializeOutputs(data []byte) TXOutputs{
	var outputs TXOutputs
	dec:= gob.NewDecoder(bytes.NewReader(data))
	err :=dec.Decode(&outputs)
	if err!=nil{
		log.Panic(err)

	}
	return outputs
}


//通过地址得到公钥的hash
func (out *TXOutput) Lock(address []byte){
	decodeAddress := Base58Decode(address)  //对公钥hash进行反系列化
	pubkeyhash := decodeAddress[1:len(decodeAddress)-4]  //去除
	out.PubkeyHash = pubkeyhash
}


//将结构体系列化
func (tx Transation) Serialize() []byte{

	var encoded bytes.Buffer
	enc := gob.NewEncoder(&encoded)

	err:=enc.Encode(tx)
	if err!=nil{
		log.Panic(err)
	}
	return encoded.Bytes()
}


//对系列化的数据进行hash,交易的hash
func (tx *Transation) Hash() []byte {
	txcopy := *tx
	txcopy.ID = []byte{}


	hash:=sha256.Sum256(txcopy.Serialize())
	return hash[:]
}

//新建一个输出
func NewTXOutput(value int, address string) *TXOutput{
	txo :=&TXOutput{value,nil}
	//txo.PubkeyHash = []byte(address)
	txo.Lock([]byte(address))
	return txo

}


//当前第一笔Coinbase交易
func NewCoinbaseTx(to,data string) *Transation  {
	//第一笔交没有TXInpu,所以设置为空 -1
	txin :=TXInput{[]byte{},-1,nil,[]byte(data)}
	txout :=NewTXOutput(subsidy,to)

	tx:=Transation{nil,[]TXInput{txin},[]TXOutput{*txout}}
	tx.ID=tx.Hash()
	return &tx

}

//在输出结构体邦定函数，判断 参数   是否在 out.PubkeyHash中 如是可以解锁
func (out *TXOutput) CanBeUnlockedWith(pubkeyhash []byte)  bool{

	return bytes.Compare(out.PubkeyHash,pubkeyhash)==0
}


//在输入结构体邦定函数，判断 参数unlockdata是否在 in.Signature中 如是可以解锁
func (in *TXInput) CanUnlockOutputWith(unlockdata []byte)  bool{
	lockinghash :=HashPubley(in.Pubkey)
	return bytes.Compare(lockinghash,unlockdata)==0
}


//判断是否为ConinBase创始区块交易
func (tx Transation) IsConinBase() bool {
	//特征：输入为1 && txid=0 &&输入所引用的输出为-1
	return len(tx.Vin) ==1 && len(tx.Vin[0].TXid)==0 && tx.Vin[0].Voutindex ==-1
}



func (tx *Transation) Sign(privkey ecdsa.PrivateKey, prevTXs map[string]Transation) {
	//先判断是不是创始交易，如是返回
	if tx.IsConinBase(){
		return
	}
	//检查过程
	for _,vin :=range tx.Vin{
		if prevTXs[hex.EncodeToString(vin.TXid)].ID == nil{
			log.Panic("ERROR:")
		}

	}

	//创建交易的副本
	txcopy := tx.TrimmedCopy()

	for inID,vin := range txcopy.Vin{
		prevTx := prevTXs[hex.EncodeToString(vin.TXid)] //得到前一笔交易的结构体
		txcopy.Vin[inID].Signature=nil
		txcopy.Vin[inID].Pubkey = prevTx.Vout[vin.Voutindex].PubkeyHash //这笔交易的输入的引用的前一笔交易输出的公钥hash
		txcopy.ID = txcopy.Hash()

		r,s,err :=ecdsa.Sign(rand.Reader,&privkey,txcopy.ID)
		if err !=nil{
			log.Panic()
		}

		signature := append(r.Bytes(),s.Bytes()...)
		tx.Vin[inID].Signature = signature
	}


}


//创建交易的副本
func (tx *Transation) TrimmedCopy() Transation {
	var inputs []TXInput
	var outputs []TXOutput
	for _,vin := range tx.Vin{
		inputs = append(inputs,TXInput{vin.TXid,vin.Voutindex,nil,nil})
	}

	for _,vout := range tx.Vout{
		outputs = append(outputs,TXOutput{vout.Value,vout.PubkeyHash})
	}

	txCopy := Transation{tx.ID,inputs,outputs}
	return txCopy
}


//验证交易
func (tx *Transation) Verify(prevTxs map[string]Transation) bool {
	//先判断是不是创始交易，如是返回
	if tx.IsConinBase(){
		return true
	}

	for _,vin := range tx.Vin{
		if prevTxs[hex.EncodeToString(vin.TXid)].ID ==nil{
			log.Panic("ERROR:")
		}
	}
	//创建交易的副本
	txcopy := tx.TrimmedCopy()

	//椭圆曲线
	curve := elliptic.P256()

	//验证
	for inID,vin := range tx.Vin{
		prevTx:= prevTxs[hex.EncodeToString(vin.TXid)]
		txcopy.Vin[inID].Signature = nil
		txcopy.Vin[inID].Pubkey = prevTx.Vout[vin.Voutindex].PubkeyHash
		txcopy.ID = txcopy.Hash()
		//初始化2个大整数
		r:=big.Int{}
		s:=big.Int{}

		siglen:= len(vin.Signature)
		r.SetBytes(vin.Signature[:(siglen/2)])
		s.SetBytes(vin.Signature[(siglen/2):])

		x:=big.Int{}
		y:=big.Int{}

		keylen :=len(vin.Pubkey)
		x.SetBytes(vin.Pubkey[:(keylen/2)])
		y.SetBytes(vin.Pubkey[(keylen/2):])

		rawPubkey := ecdsa.PublicKey{curve,&x,&y}


		if ecdsa.Verify(&rawPubkey,txcopy.ID,&r,&s) ==false{
			return false
		}

		txcopy.Vin[inID].Pubkey =nil
	}
	return true   //所有验证都不为false,验证成功返回true

}


//转账交易
func NewUTXOTransation(from,to string,amout int,bc *Blockchain) *Transation{

	var inputs []TXInput
	var outputs []TXOutput

	//通过钱包获取地址
	wallets,err := NewWallets()
	if err !=nil{
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)
	//构建输入TXInput
	//检查from地址是否有足够的金额
	acc,validoutputs := bc.FindSpendableOutputs(HashPubley(wallet.PublicKey),amout)
	//如果from地址金额小于要转账的金额，转账失败
	if acc < amout{
		log.Panic("error:Not enough funds")
	}
	//遍历输出的交易映射表
	for txid,outs := range validoutputs{
		txID,err := hex.DecodeString(txid)
		if err !=nil{
			log.Panic(err)
		}

		//遍历未花费的交易作为输入
		for _,out := range outs{

			input:=TXInput{txID,out,nil,wallet.PublicKey}
			inputs = append(inputs,input)
		}

	}


	//构建输出TXOutput
	outputs =append(outputs,*NewTXOutput(amout,to))
	//如果输入大于输出转回自己
	if acc >amout{
		outputs = append(outputs,*NewTXOutput(acc-amout,from))
	}

	//交易构建
	tx:=Transation{nil,inputs,outputs}
	//将交易的hash作为id
	tx.ID = tx.Hash()

	//添加数据签名***************************************
	bc.SignTransation(&tx,wallet.PrivateKey)

	return &tx
}

