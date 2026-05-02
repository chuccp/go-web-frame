package util

import "github.com/bytedance/sonic"

func JsonEncode(v any) (string, error) {
	marshal, err := sonic.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(marshal), nil
}
