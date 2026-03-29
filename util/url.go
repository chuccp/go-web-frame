package util

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// decodeBase64WithPadding 补齐 base64 padding 并解码
// base64 编码长度必须是 4 的倍数，不足需要补 "="
func decodeBase64WithPadding(encoded string) ([]byte, error) {
	// 计算需要补齐的 padding 数量
	if l := len(encoded) % 4; l > 0 {
		encoded += strings.Repeat("=", 4-l)
	}
	return base64.URLEncoding.DecodeString(encoded)
}

// DecodeBase64URL 解码 base64 URL 编码的字符串（自动处理 padding）
func DecodeBase64URL(encoded string) (string, error) {
	decoded, err := decodeBase64WithPadding(encoded)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// AddQueryParam adds a query parameter to a URL string.
// It automatically detects whether the URL already has query parameters
// and uses the appropriate separator (? or &).
//
// Example:
//   AddQueryParam("/path", "key", "value") -> "/path?key=value"
//   AddQueryParam("/path?foo=bar", "key", "value") -> "/path?foo=bar&key=value"
func AddQueryParam(url, key, value string) string {
	if url == "" {
		return fmt.Sprintf("?%s=%s", key, value)
	}

	separator := "?"
	for i, c := range url {
		if c == '?' && i > 0 {
			separator = "&"
			break
		}
	}

	return fmt.Sprintf("%s%s%s=%s", url, separator, key, value)
}

// AddQueryParamInt adds an integer query parameter to a URL string.
func AddQueryParamInt(url, key string, value int) string {
	return AddQueryParam(url, key, fmt.Sprintf("%d", value))
}

// AddQueryParamFlag adds a boolean flag query parameter (key=1) to a URL string.
func AddQueryParamFlag(url, key string) string {
	return AddQueryParam(url, key, "1")
}