package main

func main() {

	//TestCreateMerkleTreeRoot()
	//TestPow()
	//TestNewSerialize()
	//NewGensisBlock()
	//TestBoltDB()

	bc:=NewBlockchain("177iKmBCsDZoqJWuBARrZGau7TbbDnXkuB")
	cli:=CLI{bc}
	cli.Run()

	//wallet := NewWallet()
	//fmt.Printf("私钥：%x\n", wallet.PrivateKey.D.Bytes())
	//fmt.Printf("公钥：%x\n", wallet.PublicKey)
	//fmt.Printf("地址：%x\n", wallet.GetAddress())
	////转化16进制字节数组
	//address, _ := hex.DecodeString("31415332614e686956447a72356e41317451676b4552576b5347387261364146724c")
	//fmt.Printf("验证：%d\n", ValidateAddress(address))

}
