package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
)

const dbFile = "blockchain.db" //定义数据库常量名
const blockBucket = "blocks"   //定义了一个数据库的桶
const genesisData = "philLian blockChain"

type Blockchain struct {
	tip []byte //最近的一个区块的hash值
	db  *bolt.DB
}

type BlockChainIterateor struct {
	currenthash []byte //当前hash
	db          *bolt.DB
}

//添加区块至数据库
func (bc *Blockchain) AddBlock(block *Block){
	err :=bc.db.Update(func(tx *bolt.Tx) error {
		b:=tx.Bucket([]byte(blockBucket))
		blockIndb:=b.Get(block.Hash)
		if blockIndb !=nil{
			return nil
		}

		blockData:=block.Serialize()
		err :=b.Put(block.Hash,blockData)
		if err !=nil{
			log.Panic(err)
		}
		lastHash:=b.Get([]byte("l")) //"l"当前最高的区块
		lastBlockData:=b.Get(lastHash)
		lastblock:=DeserializeBlock(lastBlockData)
		if block.Height >lastblock.Height{
			b.Put([]byte("l"),block.Hash)
			if err !=nil{
				log.Panic(err)
			}
			bc.tip =block.Hash
		}
		return nil
	})
	if err !=nil{
		log.Panic(err)
	}
}


//添加非创始区块链至DB的操作
func (bc *Blockchain) MineBlock(transations []*Transation) *Block{

	//交易有效性验证
	for _, tx := range transations {
		if bc.VerifyTransation(tx) != true {
			log.Panic("ERROR:invalid transation!")
		} else {
			fmt.Println("Verify Success!")
		}
	}

	var lasthash []byte
	var lastheight int32
	//查询db
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockBucket))
		lasthash = b.Get([]byte("l"))
		blockdata:=b.Get(lasthash)
		block :=DeserializeBlock(blockdata)
		lastheight =block.Height
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	//新创建一个有效的区块hash
	newBlock := NewBlock(transations, lasthash,lastheight+1)

	//更新数据库
	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockBucket))
		//放入新区块的hash及系统化方法
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}
		//在桶里装数据用"l"对应当前区块的Hash
		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}
		//更新最近的一个值为最新区块hash
		bc.tip = newBlock.Hash
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return newBlock

}

//添加创始区块至DB
func NewBlockchain(address string) *Blockchain {
	var tip []byte
	//打开数据库文件
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(blockBucket))
		if b == nil {
			fmt.Println("区块链不存在，创建一个新的区块链")

			//新建一笔交易
			transation := NewCoinbaseTx(address, genesisData)
			//实例化一个创始区块
			genesis := NewGensisBlock([]*Transation{transation})
			b, err := tx.CreateBucket([]byte(blockBucket)) //创建一个桶
			if err != nil {
				log.Panic(err)
			}
			//添加创始区块至数据库
			err = b.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}
			//在桶里装数据用"l"对应当前区块的Hash
			err = b.Put([]byte("l"), genesis.Hash)
			if err != nil {
				log.Panic(err)
			}
			tip = genesis.Hash
		} else { //如果已有数据库的桶的操作程序
			tip = b.Get([]byte("l"))

		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}
	//将utxo写入至文件
	set := UTXOSet{&bc}
	set.Reindex()

	return &bc
}

//在db中实遍历查找每个区块 迭代器
func (bc *Blockchain) iterator() *BlockChainIterateor {
	bci := &BlockChainIterateor{bc.tip, bc.db}
	return bci
}

//通过当前区块获取前一区块hash
func (i *BlockChainIterateor) Next() *Block {
	var block *Block
	//查找数据库的值
	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockBucket))
		deblock := b.Get(i.currenthash)
		block = DeserializeBlock(deblock)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	i.currenthash = block.PrevBlockHash
	return block

}

//区块链打印
func (bc *Blockchain) printBlickchain() {
	bci := bc.iterator()
	for {
		block := bci.Next()
		block.String()
		fmt.Println()
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

//在区块中找到未花费的输出，表示可以引用的输出（可以用的花费）
func (bc *Blockchain) FindUnspentTransations(pubkeyhash []byte) []Transation {
	//声明返回的变量，所有未花费的变量
	var unspentTxs []Transation

	//存储已花费的交易
	spendTXOs := make(map[string][]int) //string交易的hash -> []int输出已花费序列号 映射关系表

	//循环遍历每个区块(从后往前) 迭代器
	bci := bc.iterator()
	for {
		block := bci.Next()

		//循环遍历每个区块的交易 tx第笔交易
		for _, tx := range block.Transations {

			txID := hex.EncodeToString(tx.ID) //将交易hash转成string

			//循环遍历每个每笔交易的输出是否有被花费
		output:
			for outIdx, out := range tx.Vout { //outIdx输出系号
				if spendTXOs[txID] != nil {

					//循环遍历所有输出的交易记录
					for _, spentOut := range spendTXOs[txID] {

						//如果遍历每笔交易等于outIdx输出系号表示已被花费
						if spentOut == outIdx {
							continue output
						}
					}
				}
				//将可以解锁的交易布储至unspentxs
				if out.CanBeUnlockedWith(pubkeyhash) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}
			// 如果这笔交易是coinbase
			if tx.IsConinBase() == false {
				//循环遍历每个每笔交易的输入
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(pubkeyhash) {
						inTxID := hex.EncodeToString(in.TXid)
						spendTXOs[inTxID] = append(spendTXOs[inTxID], in.Voutindex)
					}
				}
			}
		}

		//如果当前区块的前一区块的长度为0，表示已到创始区块结束循环
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return unspentTxs
}

//遍历UTXO可花费的输出
func (bc *Blockchain) FindUTXO(pubkeyhash []byte) []TXOutput {
	var UTXOs []TXOutput
	unspendTransations := bc.FindUnspentTransations(pubkeyhash)
	for _, tx := range unspendTransations {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(pubkeyhash) {
				UTXOs = append(UTXOs, out)
			}

		}
	}
	return UTXOs
}

//看一下是否有足够的金额可以转账
func (bc *Blockchain) FindSpendableOutputs(pubkeyhash []byte, amout int) (int, map[string][]int) {
	//建立一张映射表
	unspentOutputs := make(map[string][]int)
	//获得未花费的金额
	unspentTXs := bc.FindUnspentTransations(pubkeyhash)
	//初始化总金额
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)
		//查找每个交易当中的未花费输出
		for outIdx, out := range tx.Vout {

			if out.CanBeUnlockedWith(pubkeyhash) && accumulated < amout {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				//如果要输出金额大于等于输入金额
				if accumulated >= amout {
					break Work
				}
			}
		}
	}
	//返回要输出的总金额及输出的交易映射表
	return accumulated, unspentOutputs

}

//添加数据签名
func (bc *Blockchain) SignTransation(tx *Transation, prikey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transation) //所有的输入构建映射表
	//遍历输入的输出
	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransationById(vin.TXid) //找到前一笔交易的输出
		if err != nil {
			log.Panic(err)
		}

		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX

	}
	tx.Sign(prikey, prevTXs)
}

//查找交易的id是否存在（找到前一笔交易的输出）
func (bc *Blockchain) FindTransationById(ID []byte) (Transation, error) {
	bci := bc.iterator()
	for {
		block := bci.Next()
		for _, tx := range block.Transations {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return Transation{}, errors.New("transation is not find!")
}

//验证数据签名
func (bc *Blockchain) VerifyTransation(tx *Transation) bool {
	//引用的所有输出
	prevTXs := make(map[string]Transation)

	//遍历输入的输出
	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransationById(vin.TXid)
		if err != nil {
			log.Panic(err)
		}

		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX

	}

	return tx.Verify(prevTXs)
}

//查找所有未花费的输出
func (bc *Blockchain) FindAllUTXO() map[string]TXOutputs {
	UTXO := make(map[string]TXOutputs) //输出
	spentTXs := make(map[string][]int) //输入
	bci := bc.iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transations {
			txID := hex.EncodeToString(tx.ID)
		Outputs:
			for outIdx, out := range tx.Vout {
				//如果不为空表示存在于已花费的输出中
				if spentTXs[txID] != nil {
					for _, spentOutIds := range spentTXs[txID] {
						if spentOutIds == outIdx {
							continue Outputs
						}
					}
				}
				//存储未花费的输出
				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				//更新map表
				UTXO[txID] = outs
			}
			if tx.IsConinBase() == false {
				for _, in := range tx.Vin {
					inTXID := hex.EncodeToString(in.TXid)
					spentTXs[inTXID] = append(spentTXs[inTXID], in.Voutindex)
				}
			}
		}

		//如果上一个区块hash为空表这这是创始区块退出
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return UTXO
}

//获取最新区块高度
func (bc *Blockchain) GetBestHeight() int32  {
	var lastBlock Block
	err :=bc.db.View(func(tx *bolt.Tx) error {
		b:=tx.Bucket([]byte(blockBucket))
		lastHash := b.Get([]byte("l")) //获取最后区块hash
		blockdata := b.Get(lastHash)
		lastBlock = *DeserializeBlock(blockdata)
		return nil

	})
	if err !=nil{
		log.Panic(err)
	}
	return lastBlock.Height
}



//获取最新区块hash值
func (bc *Blockchain) GetBlockhash() [][]byte {
	var blocks [][]byte
	bci:=bc.iterator()
	for{
		block:=bci.Next()
		blocks=append(blocks,block.Hash)

		if len(block.PrevBlockHash)==0{
			break
		}
	}
	return blocks
}


//获取最新区块
func (bc *Blockchain) GetBlock(blockHash []byte) (Block, error) {
	var block Block
	err :=bc.db.View(func(tx *bolt.Tx) error {
		b:=tx.Bucket([]byte(blockBucket))
		blcokData:=b.Get(blockHash)
		if blcokData == nil{
			return errors.New("Block is not Fund")
		}
		block = *DeserializeBlock(blcokData)
		return nil
	})
	if err !=nil{
		return block,err
	}

	return block,nil

}