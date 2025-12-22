package util

import "hash/crc32"

func CRC(len int, key string) string {
	code := GenerateRandomStringByAlphanumeric(len)
	hashInput := code + key
	hash := crc32.ChecksumIEEE([]byte(hashInput))
	first := int(hash%26+26) % 26
	second := int((hash>>8)%26+26) % 26
	return code + Index2Str(first, Uppercase) + Index2Str(second, Uppercase)
}
