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
//
//	AddQueryParam("/path", "key", "value") -> "/path?key=value"
//	AddQueryParam("/path?foo=bar", "key", "value") -> "/path?foo=bar&key=value"
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

// isWindowsDrivePath checks if the path starts with a Windows drive letter (like C:/)
func isWindowsDrivePath(path string) bool {
	if len(path) < 2 {
		return false
	}
	// Check for pattern like "C:" or "C/"
	return ((path[0] >= 'A' && path[0] <= 'Z') || (path[0] >= 'a' && path[0] <= 'z')) && path[1] == ':'
}
// JoinUrl joins URL path segments into a single URL string.
// It properly handles leading and trailing slashes between segments.
// It also normalizes backslashes to forward slashes (useful for Windows paths).
//
// Example:
//
//	JoinUrl("http://example.com", "api", "users") -> "http://example.com/api/users"
//	JoinUrl("/api", "users", "1") -> "/api/users/1"
//	JoinUrl("http://example.com/", "/api/", "/users/") -> "http://example.com/api/users"
//	JoinUrl("C:\\static\\voice", "58\\file.mp3") -> "C:/static/voice/58/file.mp3"
func JoinUrl(root string, args ...string) string {
	if root == "" && len(args) == 0 {
		return ""
	}

	// Collect all parts
	parts := make([]string, 0, len(args)+1)
	if root != "" {
		parts = append(parts, root)
	}
	parts = append(parts, args...)

	// Find first non-empty part
	firstIdx := -1
	for i, part := range parts {
		if part != "" {
			firstIdx = i
			break
		}
	}
	if firstIdx == -1 {
		return "/"
	}

	// Check if first part has a scheme (like http:// or https://)
	hasScheme := false
	if len(parts[firstIdx]) > 7 {
		prefix := strings.ToLower(parts[firstIdx][:7])
		hasScheme = strings.HasPrefix(prefix, "http://") || strings.HasPrefix(prefix, "https:/")
	}

	var result strings.Builder
	for i, part := range parts {
		if part == "" {
			continue
		}

		// Convert backslashes to forward slashes (handle Windows paths)
		part = strings.ReplaceAll(part, "\\", "/")

		// For the first non-empty part (root), preserve leading slashes and scheme
		if i == firstIdx {
			// Remove trailing slash from root
			part = strings.TrimRight(part, "/")
			result.WriteString(part)
			continue
		}

		// Remove leading and trailing slashes from other parts
		part = strings.Trim(part, "/")
		if part == "" {
			continue
		}

		result.WriteString("/")
		result.WriteString(part)
	}

	urlStr := result.String()

	// Handle empty result
	if urlStr == "" {
		return "/"
	}

	// If no scheme and doesn't start with /, add leading slash for path-style URL
	// But don't add leading slash for Windows drive paths (like C:/)
	if !hasScheme && !strings.HasPrefix(urlStr, "/") && !isWindowsDrivePath(urlStr) {
		urlStr = "/" + urlStr
	}

	// For scheme URLs, ensure we don't break them
	if hasScheme && !strings.Contains(urlStr[7:], "/") {
		// URL like "http://example.com" - no path, keep as is
		return urlStr
	}

	return urlStr
}
