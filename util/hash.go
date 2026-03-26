package util

import (
	"crypto/sha1"
	"encoding/hex"
)

// SHA1 计算字符串的 SHA1 哈希值并返回十六进制编码结果
func SHA1(str string) string {
	h := sha1.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// SHA1Bytes 计算字节数组的 SHA1 哈希值并返回十六进制编码结果
func SHA1Bytes(data []byte) string {
	h := sha1.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}