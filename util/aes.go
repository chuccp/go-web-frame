package util

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
)

// EncryptByCBC AES-256-CBC 加密实现
func EncryptByCBC(text string, key string, iv string) string {
	// 校验密钥长度（AES-256 要求密钥长度为 32 字节）
	if len(key) != 32 {
		panic(errors.New("AES-256 密钥长度必须为 32 字节"))
	}
	// 校验IV长度（CBC模式要求IV长度等于块大小，AES块大小固定为16字节）
	if len(iv) != 16 {
		panic(errors.New("CBC模式IV长度必须为16字节"))
	}

	// 将明文转换为字节数组
	plaintext := []byte(text)

	// 创建AES密码块
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		panic(err)
	}

	// 使用PKCS#7填充明文（确保长度为块大小的整数倍）
	padding := aes.BlockSize - len(plaintext)%aes.BlockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	plaintext = append(plaintext, padtext...)

	// 创建CBC模式的加密流
	mode := cipher.NewCBCEncrypter(block, []byte(iv))

	// 执行加密（输出与输入长度相同）
	ciphertext := make([]byte, len(plaintext))
	mode.CryptBlocks(ciphertext, plaintext)

	// 加密结果使用Base64编码返回（便于传输和存储）
	return base64.StdEncoding.EncodeToString(ciphertext)
}

// DecryptByCBC AES-256-CBC 解密实现
func DecryptByCBC(cipherText string, key string, iv string) string {
	// 校验密钥和IV长度（与加密保持一致）
	if len(key) != 32 {
		panic(errors.New("AES-256 密钥长度必须为 32 字节"))
	}
	if len(iv) != 16 {
		panic(errors.New("CBC模式IV长度必须为16字节"))
	}

	// 先对密文进行Base64解码（加密时做了Base64编码）
	ciphertext, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		panic(err)
	}

	// 创建AES密码块
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		panic(err)
	}

	// 检查密文长度是否为块大小的整数倍（解密要求）
	if len(ciphertext)%aes.BlockSize != 0 {
		panic(errors.New("密文长度必须是16字节的整数倍"))
	}

	// 创建CBC模式的解密流
	mode := cipher.NewCBCDecrypter(block, []byte(iv))

	// 执行解密（输出与输入长度相同）
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// 去除PKCS#7填充（加密时添加的填充）
	padding := int(plaintext[len(plaintext)-1])
	if padding < 1 || padding > aes.BlockSize {
		panic(errors.New("无效的填充数据"))
	}
	plaintext = plaintext[:len(plaintext)-padding]

	// 转换为字符串返回
	return string(plaintext)
}
