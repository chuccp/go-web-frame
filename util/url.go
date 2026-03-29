package util

import "fmt"

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