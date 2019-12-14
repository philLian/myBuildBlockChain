package main

//客户端与用户交互接口定制


import (
	"flag"
	"fmt"
	"log"
	"os"
)

type CLI struct {
	bc *Blockchain
}

func (cli *CLI) addBlock()  {
	//调用方法 AddBlock() 增加一个区块
	cli.bc.MineBlock([]*Transation{})
}

func (cli *CLI) printChain()  {
	//调用方法printChain() 增加一个区块
	cli.bc.printBlickchain()
}

func (cli *CLI) getBalance(address string)  {
	// 获取UTXOs金额
	balance :=0
	decodeAddress :=Base58Decode([]byte(address))
	pubkeyhash :=decodeAddress[1:len(decodeAddress)-4]

	set :=UTXOSet{cli.bc}
	//UTXOs := cli.bc.FindUTXO(pubkeyhash)
	UTXOs := set.FindUTXObyPubkeyHash(pubkeyhash)
	for _,out := range UTXOs{
		balance +=out.Value
	}

	fmt.Printf("balance of '%s':%d\n",address,balance)
}

func (cli *CLI) send(from,to string,amount int)  {
	// 转账交易方法
	//构建一笔交易
	tx:=NewUTXOTransation(from,to,amount,cli.bc)
	//挖矿操作
	newblock :=cli.bc.MineBlock([]*Transation{tx})
	set :=UTXOSet{cli.bc}
	set.update(newblock)
	cli.getBalance("1HURQgxhhtdnCauPxjx8wjs7KvMxPZ5tnB")
	cli.getBalance("1MR7VmbX2u2Y4Ti76oDNM4tKR1zfBvCk7K")
	cli.getBalance("177iKmBCsDZoqJWuBARrZGau7TbbDnXkuB")
	cli.getBalance("1kHxDMSi166Uiq1JszmfLo3t7jo4Yjp1i")
	cli.getBalance("1P3eN3bTMewgZY9H3MRAFc1Vzu8r3TKQyY")

	fmt.Printf("Success!")

}


func (cli *CLI) createWallet()  {
	//调用方法NewWallet()
	wallets,_:=NewWallets()
	//if err!=nil{
	//	log.Panic(err)
	//}

	adderss :=wallets.CreateWallet()
	wallets.SaveToFile()
	fmt.Printf("your address:%s\n",adderss)
}

func (cli *CLI) listAddress()  {
	//调用方法NewWallet()
	wallets,err:=NewWallets()
	if err!=nil{
		log.Panic(err)
	}

	adderss :=wallets.GetAddress()
	for _,adderss :=range adderss{
		fmt.Println(adderss)
	}
}







//如果输入小于1退出
func (cli *CLI)validateArgs(){
	if len(os.Args) <1{
		fmt.Println("参烤小于1")
		os.Exit(1)
	}
	fmt.Println(os.Args)
}

//信息提示
func (cli *CLI) printUsage() {
	fmt.Println("USages:")
	fmt.Println("addblock-增加区块")
	fmt.Println("printChain-打印区块链")
	fmt.Println("createwallet-创建钱包")
	fmt.Println("listaddress-钱包列表")
}


func (cli *CLI) Run(){
	//查看是否小于1
	cli.validateArgs()

	nodeID :=os.Getenv("NODE_ID")  //获取系统环境变量
	println(nodeID)
	if nodeID ==""{

		fmt.Printf("NODE_ID is not set! ")
		os.Exit(1)
	}




	//生成交互指针
	addBlockCmd :=flag.NewFlagSet("addblock",flag.ExitOnError)
	printChainCmd :=flag.NewFlagSet("printChain",flag.ExitOnError)

	getBalanceCmd :=flag.NewFlagSet("getbalance",flag.ExitOnError)
	getBalanceAddress:=getBalanceCmd.String("address","","the address to get balance of")

	sendCmd := flag.NewFlagSet("send",flag.ExitOnError)
	sendFrom :=sendCmd.String("from","","source walllet address")
	sendTo :=sendCmd.String("to","","Destination wallet address")
	sendAmount :=sendCmd.Int("amount",0,"Amount to send")

	createWalletCMD :=flag.NewFlagSet("createwallet",flag.ExitOnError)
	listAddressCMD :=flag.NewFlagSet("listaddress",flag.ExitOnError)

	getBastHeightCmd :=flag.NewFlagSet("getBastHeight",flag.ExitOnError)

	startNodeCmd :=flag.NewFlagSet("startNode",flag.ExitOnError)
	startNodeMinner :=startNodeCmd.String("minner","","minner address")

	//对输入命令进行解析
	switch os.Args[1] {
	case "startNode":
		err:=startNodeCmd.Parse(os.Args[2:])
		if err!=nil{
			log.Panic(err)
		}



	//如解析“addblock”
	case "addblock":
		err:=addBlockCmd.Parse(os.Args[2:])
		if err!=nil{
			log.Panic(err)
		}
	case "printChain":
		err:=printChainCmd.Parse(os.Args[2:])
		if err!=nil{
			log.Panic(err)
		}

	case "getbalance":
		err:=getBalanceCmd.Parse(os.Args[2:])
		if err!=nil{
			log.Panic(err)
		}
	case "send":
		err:=sendCmd.Parse(os.Args[2:])
		if err!=nil{
			log.Panic(err)
		}

	case "createwallet":
		err:=createWalletCMD.Parse(os.Args[2:])
		if err!=nil{
			log.Panic(err)
		}

	case "listaddress":
		err:=listAddressCMD.Parse(os.Args[2:])
		if err!=nil{
			log.Panic(err)
		}

	case "getBastHeight":
		err:=getBastHeightCmd.Parse(os.Args[2:])
		if err!=nil{
			log.Panic(err)
		}

	default:
		cli.printUsage()
		os.Exit(1) //退出
	}



	//解析成功执行方法
	if addBlockCmd.Parsed(){
		cli.addBlock()
	}

	if printChainCmd.Parsed(){
		cli.printChain()
	}

	if getBalanceCmd.Parsed(){
		if *getBalanceAddress ==""{
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}

	if sendCmd.Parsed(){
		if *sendFrom =="" || *sendTo =="" || *sendAmount <= 0{
			os.Exit(1)
		}
		cli.send(*sendFrom,*sendTo,*sendAmount)
	}


	if createWalletCMD.Parsed(){
		cli.createWallet()
	}

	if listAddressCMD.Parsed(){
		cli.listAddress()
	}

	if getBastHeightCmd.Parsed(){

		cli.getBastHeight()
	}

	if startNodeCmd.Parsed(){
		nodeID:=os.Getenv("NODE_ID")
		if nodeID ==""{
			startNodeCmd.Usage()
			os.Exit(1)
		}
		cli.startNode(nodeID,*startNodeMinner)
	}
}


//开启服务
func (cli *CLI) startNode(nodeID string, minnerAddress string) {

	fmt.Printf("Starting node:%s\n",nodeID)
	if len(minnerAddress)>0{
		if ValidateAddress([]byte(minnerAddress)){
			fmt.Println("%s-minner is on",minnerAddress)
		}else {
			log.Panic("error minner Address")
		}
	}

	StartServer(nodeID,minnerAddress,cli.bc)
}





//区块链最新高度
func (cli *CLI) getBastHeight()  {
	fmt.Println(cli.bc.GetBestHeight())
}
