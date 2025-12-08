package db

import (
	"log"
	"net/url"
	"testing"
)

func TestName(t *testing.T) {
	//newURL := &url.URL{
	//	Scheme: "https",
	//	Host:   "api.example.com",
	//	Path:   "/user/profile",
	//}

	url_, err := url.Parse("mysql://xxx:zxxx@localhost:8080")
	log.Println(url_, err)
}
