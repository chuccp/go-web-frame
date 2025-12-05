package util

import (
	"net"
	"net/url"
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
