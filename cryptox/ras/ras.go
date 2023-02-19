package ras

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
)

// GenerateRSAKey 生成RSA私钥和公钥，保存到文件中
// bits 证书大小
func GenerateRSAKey(bits int) (private string, public string, err error) {
	pemBuf := make([]byte, 0)
	pubBuf := make([]byte, 0)
	//GenerateKey函数使用随机数据生成器random生成一对具有指定字位数的RSA密钥
	//Reader是一个全局、共享的密码用强随机数生成器
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", "", err
	}
	//保存私钥
	//通过x509标准将得到的ras私钥序列化为ASN.1 的 DER编码字符串
	X509PrivateKey := x509.MarshalPKCS1PrivateKey(privateKey)
	//使用pem格式对x509输出的内容进行编码
	//创建文件保存私钥

	privateBuf := bytes.NewBuffer(pemBuf)

	//构建一个pem.Block结构体对象
	privateBlock := pem.Block{Type: "RSA Private Key", Bytes: X509PrivateKey}
	//将数据保存到文件
	err = pem.Encode(privateBuf, &privateBlock)
	if err != nil {
		return "", "", err
	}

	publicKey := privateKey.PublicKey
	//X509对公钥编码
	X509PublicKey, err := x509.MarshalPKIXPublicKey(&publicKey)
	if err != nil {
		return "", "", err
	}
	publicBuf := bytes.NewBuffer(pubBuf)
	publicBlock := pem.Block{Type: "RSA Public Key", Bytes: X509PublicKey}
	err = pem.Encode(publicBuf, &publicBlock)
	if err != nil {
		return "", "", err
	}
	return privateBuf.String(), publicBuf.String(), nil
}

// EncryptRSA RSA加密
// plainText 要加密的数据
// path 公钥匙文件地址
func EncryptRSA(plainText string, pubKey string) (cipher string, err error) {

	//pem解码
	block, _ := pem.Decode([]byte(pubKey))
	//x509解码

	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}
	//类型断言
	publicKey := publicKeyInterface.(*rsa.PublicKey)
	//对明文进行加密
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, []byte(plainText))
	if err != nil {
		return "", err
	}
	//返回密文
	cipher = base64.StdEncoding.EncodeToString(cipherText)
	return cipher, nil
}

// DecryptRSA RSA解密
// cipherText 需要解密的byte数据
// path 私钥文件路径
func DecryptRSA(cipherText string, pemKey string) (plain string, err error) {
	cipherByte, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}
	//pem解码
	block, _ := pem.Decode([]byte(pemKey))
	//X509解码
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}
	//对密文进行解密
	plainText, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, cipherByte)
	if err != nil {
		return "", err
	}
	//返回明文
	plain = string(plainText)
	return plain, nil
}
