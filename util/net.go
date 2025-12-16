package util

import (
	"net"
	"net/url"
	"regexp"
)

func GetDomainFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return ""
	}
	host, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		return u.Host
	}
	return host
}
func GetDomainFromHost(host string) string {
	host_, _, err := net.SplitHostPort(host)
	if err != nil {
		return host
	}
	return host_
}
func GetHost(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return ""
	}
	return u.Host
}
func IsDomain(str string) bool {
	// 定义域名的正则表达式
	domainRegex := `^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(domainRegex, str)
	return matched
}
