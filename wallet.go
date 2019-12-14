package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"golang.org/x/crypto/ripemd160"
)

//声明版本号常量
const version = byte(0x00)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

//新建钱包（其实就是生成一个私钥和公钥）
func NewWallet() *Wallet {
	private, public := newKeyPair()

	wallet := Wallet{private, public}

	return &wallet
}

//生成私钥和公钥
func newKeyPair() (ecdsa.PrivateKey, []byte) {
	//生成椭圆曲线 secp256r1曲线  比特币当中的曲是secp256k1
	curve := elliptic.P256()

	//随机数的种子生生随机数 生成个公钥及私钥个结构体，存储了公钥
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		fmt.Println("error")
	}

	//公钥生成
	pubkey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, pubkey
}

//帮定钱包封装地址生成函数
func (w Wallet) GetAddress() []byte {
	//对pubkey进行 hash
	pubkeyHash := HashPubley(w.PublicKey)

	//合并版本号及公钥hash
	versionPayload := append([]byte{0x00}, pubkeyHash...)
	//计算校验码
	check := checksum(versionPayload)
	//生成完整的地址{版本号+公钥+校校码}
	fullPayload := append(versionPayload, check...)

	//对fullPayload进行base58编码
	address := Base58Encode(fullPayload)
	return address
}

//公钥hash函数
func HashPubley(pubkey []byte) []byte {

	//对pubkey进行 hash256
	pubkeyHash256 := sha256.Sum256(pubkey)
	//实体化一个ripemd160的对象
	PIPEMD160Hash := ripemd160.New()

	//对pubkeyHash256进行PIPEMD160Hash计算
	_, err := PIPEMD160Hash.Write(pubkeyHash256[:])
	//如果不为空，表示出错失败
	if err != nil {
		fmt.Println("error")

	}

	//对pubkeyHash256进行PIPEMD160Hash计算
	publicRIPEMD160 := PIPEMD160Hash.Sum(nil)
	return publicRIPEMD160

}

//计算校验码
func checksum(payload []byte) []byte {
	//对参数进行双hash256操作取前4个字节
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])
	//取前4字节
	checksum := secondSHA[:4]

	return checksum
}

//钱包地址验证
func ValidateAddress(address []byte) bool {
	//对address解码
	pubkeyHash := Base58Decode(address)
	//获取校码码最后四位
	actualCheckSum := pubkeyHash[len(pubkeyHash)-4:]
	//获取公钥hash
	publickeyHash := pubkeyHash[1 : len(pubkeyHash)-4]
	//从新hash
	targetChecksum := checksum(append([]byte{0x00}, publickeyHash...))

	//对比两个数完全相同==0 返回True
	return  bytes.Compare(actualCheckSum,targetChecksum)==0



}
