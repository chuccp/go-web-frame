package util

import "testing"

func TestEncryptByCBC(t *testing.T) {

	key := "12345678901234567890123456789012"
	iv := "1234567890123456"
	text := "hello world"
	cipherText, _ := EncryptByCBC(text, key, iv)
	plainText, _ := DecryptByCBC(cipherText, key, iv)
	if plainText != text {
		t.Errorf("EncryptByCBC failed, expected %s, got %s", text, plainText)
	} else {
		t.Logf("EncryptByCBC success, cipherText: %s", cipherText)
	}

}
