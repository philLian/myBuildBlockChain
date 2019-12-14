package main

//UTXO数据库读写操作

import (
	"encoding/hex"
	"github.com/boltdb/bolt"
	"log"
)

type UTXOSet struct {
	bchain *Blockchain
}

const utxoBucket = "chainset"

//重置数据库并写入桶
func (u UTXOSet) Reindex() {
	db := u.bchain.db
	bucketName := []byte(utxoBucket)

	//删除桶并新建
	err := db.Update(func(tx *bolt.Tx) error {
		err2 := tx.DeleteBucket(bucketName)
		if err2 != nil && err2 != bolt.ErrBucketNotFound {
			log.Panic(err2)
		}
		_, err3 := tx.CreateBucket(bucketName)
		if err3 != nil {
			log.Panic(err3)
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	UTXO := u.bchain.FindAllUTXO()
	err4 := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		//循环所有的UTXO map 表
		for txID, outs := range UTXO {
			key, err5 := hex.DecodeString(txID)
			if err5 != nil {
				log.Panic(err5)
			}

			err6 := b.Put(key, outs.Serializa())
			if err6 != nil {
				log.Panic(err6)
			}
		}
		return nil
	})
	if err4 != nil {
		log.Panic(err4)
	}
}

//pubkeyhash所有未花费的输出
func (u UTXOSet) FindUTXObyPubkeyHash(pubkeyhash []byte) []TXOutput {
	var UTXOs []TXOutput
	db := u.bchain.db
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs := DeserializeOutputs(v)
			for _, out := range outs.Outputs {
				if out.CanBeUnlockedWith(pubkeyhash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return UTXOs
}

//新建区块更新UTXO
func (u UTXOSet) update(block *Block) {
	db := u.bchain.db
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))

		for _, tx := range block.Transations {
			updateouts := TXOutputs{}
			if tx.IsConinBase() == false {
				for _, vin := range tx.Vin {
					outsbytes := b.Get(vin.TXid)
					outs := DeserializeOutputs(outsbytes)
					for outIdx, out := range outs.Outputs {
						if outIdx != vin.Voutindex {
							updateouts.Outputs = append(updateouts.Outputs, out)
						}
					}

					if len(updateouts.Outputs) == 0 {
						err := b.Delete(vin.TXid)
						if err != nil {
							log.Panic(err)
						}
					} else {
						err := b.Put(vin.TXid, updateouts.Serializa())
						if err != nil {
							log.Panic(err)
						}
					}
				}
			}
			newOutputs := TXOutputs{}
			for _, out := range tx.Vout {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}
			err := b.Put(tx.ID, newOutputs.Serializa())
			if err != nil {
				log.Panic(err)
			}
		}
		return nil

	})
	if err != nil {
		log.Panic(err)
	}

}
