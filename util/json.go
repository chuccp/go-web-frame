package util

import "encoding/json"

func JsonEncode(v any) (string, error) {
	marshal, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(marshal), nil
}
