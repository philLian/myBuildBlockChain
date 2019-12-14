package main

import (
	"bytes"
	"math/big"
)

//切片存储base58的元素字母
var B58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

//字节数组编码
func Base58Encode(input []byte) []byte {

	var result []byte //声明一个字节切片，这是一个反回值

	//初始为为一个大整数，开始为0，然后设置为input
	x := big.NewInt(0).SetBytes(input)
	base := big.NewInt(int64(len(B58Alphabet)))
	zero := big.NewInt(0)

	//这是一个大整数的指针
	mod := &big.Int{}

	//x的结果不为0时执行取余操作，大小为58进制
	for x.Cmp(zero) != 0 {
		x.DivMod(x, base, mod)                            //取x的余数
		result = append(result, B58Alphabet[mod.Int64()]) //取余数的字符添加到result
	}

	ReverseBytes(result)	//将得到的结果进行反转
	//如果结果的前面为"0"换成"1"
	for _, b:=range input{
		if b == 0x00{
			result = append([]byte{B58Alphabet[0]},result...)
		}else {
			break
		}
	}


	return result

}

//字节数组解码
func Base58Decode(input []byte) []byte{

	//初始为为一个大整数，开始为0，
	result := big.NewInt(0)
	zeroBytes :=0
	//如果前面为为返转为1,计算有几个1，给下一步反转0
	for _,b:=range input{
		if b=='1'{
			zeroBytes++
		}else{
			break
		}
	}

	//除去前面的 1之后开始操作
	payload:= input[zeroBytes:]

	//循环逆推出结果
	for _, b :=range payload{
		charIndex := bytes.IndexByte(B58Alphabet,b) //反推出余数
		result.Mul(result,big.NewInt(58))	//将之前的结果*58
		result.Add(result,big.NewInt(int64(charIndex))) //然后+余数

	}


	//将大整数转化为字节数组
	decoded:=result.Bytes()
	//对编码0转为1时，转回0
	decoded = append(bytes.Repeat([]byte{0x00},zeroBytes),decoded...)

	return decoded

}


