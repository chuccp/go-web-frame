package util

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"

	"emperror.dev/errors"
)

// EncryptByCBC AES-256-CBC 加密实现
func EncryptByCBC(text string, key string, iv string) (string, error) {
	// 校验密钥长度（AES-256 要求密钥长度为 32 字节）
	if len(key) != 32 {
		return "", errors.New("The key length of AES-256 must be 32 bytes.")
	}
	// 校验IV长度（CBC模式要求IV长度等于块大小，AES块大小固定为16字节）
	if len(iv) != 16 {
		return "", errors.New("The IV length in CBC mode must be 16 bytes.")
	}

	// 将明文转换为字节数组
	plaintext := []byte(text)

	// 创建AES密码块
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", errors.WithStackIf(err)
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
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// DecryptByCBC AES-256-CBC 解密实现
func DecryptByCBC(cipherText string, key string, iv string) (string, error) {
	// 校验密钥和IV长度（与加密保持一致）
	if len(key) != 32 {
		return "", errors.New("The key length of AES-256 must be 32 bytes")
	}
	if len(iv) != 16 {
		return "", errors.New("The IV length in CBC mode must be 16 bytes")
	}

	// 先对密文进行Base64解码（加密时做了Base64编码）
	ciphertext, err := base64.URLEncoding.DecodeString(cipherText)
	if err != nil {
		return "", errors.WithStackIf(err)
	}

	// 创建AES密码块
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", errors.WithStackIf(err)
	}

	// 检查密文长度是否为块大小的整数倍（解密要求）
	if len(ciphertext)%aes.BlockSize != 0 {
		return "", errors.New("The ciphertext length must be an integer multiple of 16 bytes")
	}

	// 创建CBC模式的解密流
	mode := cipher.NewCBCDecrypter(block, []byte(iv))

	// 执行解密（输出与输入长度相同）
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// 去除PKCS#7填充（加密时添加的填充）
	padding := int(plaintext[len(plaintext)-1])
	if padding < 1 || padding > aes.BlockSize {
		return "", errors.New("Invalid padding data")
	}
	plaintext = plaintext[:len(plaintext)-padding]
	// 转换为字符串返回
	return string(plaintext), nil
}

// DecryptCBCBytes AES-256-CBC 解密（字节数组版本）
// key: 32字节密钥
// iv: 16字节IV（如果为nil则使用key的前16字节）
// ciphertext: 密文
// 返回：解密后的明文
func DecryptCBCBytes(key, iv, ciphertext []byte) ([]byte, error) {
	// 校验密钥长度
	if len(key) != 32 {
		return nil, errors.New("The key length of AES-256 must be 32 bytes")
	}

	// 如果IV为空，使用key的前16字节
	if iv == nil {
		iv = key[:16]
	}

	// 校验IV长度
	if len(iv) != 16 {
		return nil, errors.New("The IV length in CBC mode must be 16 bytes")
	}

	// 检查密文长度是否为块大小的整数倍
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, errors.New("The ciphertext length must be an integer multiple of 16 bytes")
	}

	// 创建AES密码块
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.WithStackIf(err)
	}

	// 创建CBC模式的解密流
	mode := cipher.NewCBCDecrypter(block, iv)

	// 执行解密
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// 去除PKCS#7填充
	plaintext, err = PKCS7Unpad(plaintext)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// DecryptCBCBase64 AES-256-CBC 解密（Base64密文版本）
// key: 32字节密钥
// iv: 16字节IV（如果为nil则使用key的前16字节）
// cipherTextBase64: Base64编码的密文
// 返回：解密后的明文
func DecryptCBCBase64(key, iv []byte, cipherTextBase64 string) ([]byte, error) {
	// Base64解码密文
	ciphertext, err := base64.StdEncoding.DecodeString(cipherTextBase64)
	if err != nil {
		return nil, errors.WithStackIf(err)
	}

	return DecryptCBCBytes(key, iv, ciphertext)
}

// PKCS7Unpad 去除 PKCS7 填充
func PKCS7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("data is empty")
	}

	padding := int(data[len(data)-1])
	if padding < 1 || padding > aes.BlockSize {
		return nil, errors.New("invalid padding value")
	}

	// 验证填充
	for i := len(data) - padding; i < len(data); i++ {
		if int(data[i]) != padding {
			return nil, errors.New("invalid padding")
		}
	}

	return data[:len(data)-padding], nil
}

// PKCS7Pad 添加 PKCS7 填充
func PKCS7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}
