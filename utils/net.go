package utils

import (
	"crypto/tls"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var defHeaders = make(map[string]string)

func init() {
	defHeaders["User-Agent"] = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.12; rv:52.0) Gecko/20100101 Firefox/52.0"
	defHeaders["Accept-Language"] = "ru-RU,ru;q=0.8,en-US;q=0.5,en;q=0.3"
	defHeaders["Referer"] = "https://ya.ru/"
	defHeaders["Cookie"] = ""
}

// Config for http dialer
type Config struct {
	ConnectTimeout   time.Duration
	ReadWriteTimeout time.Duration
}

// HTTPImgLen return len of image by url or 0
func HTTPImgLen(url string) int64 {
	client := NewTimeoutClient()
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		log.Println(err)
		return 0
	}
	for k, v := range defHeaders {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return 0
	}
	if strings.HasPrefix(resp.Header.Get("Content-Type"), "image") {
		return resp.ContentLength
	}
	return 0

}

// NewTimeoutClient - create http client with TimeOut and disabled http/2
func NewTimeoutClient(args ...interface{}) *http.Client {
	// Default configuration
	config := &Config{
		ConnectTimeout:   5 * time.Second,
		ReadWriteTimeout: 5 * time.Second,
	}

	// merge the default with user input if there is one
	if len(args) == 1 {
		timeout := args[0].(time.Duration)
		config.ConnectTimeout = timeout
		config.ReadWriteTimeout = timeout
	}

	if len(args) == 2 {
		config.ConnectTimeout = args[0].(time.Duration)
		config.ReadWriteTimeout = args[1].(time.Duration)
	}
	http.DefaultTransport.(*http.Transport).TLSNextProto = make(map[string]func(string, *tls.Conn) http.RoundTripper)

	return &http.Client{
		Transport: &http.Transport{
			Dial: TimeoutDialer(config),
		},
	}
}

// TimeoutDialer try Dial/ReadWrite in Timeout
func TimeoutDialer(config *Config) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, config.ConnectTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(config.ReadWriteTimeout))
		return conn, nil
	}
}

// HTTPGetBody create get request with default headers
// return nil or data
func HTTPGetBody(url string) []byte {

	client := NewTimeoutClient()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil
	}
	for k, v := range defHeaders {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		return body
	}

	return nil
}

// checkAndCreate may create dirs
func СheckAndCreate(path string) (bool, error) {
	// detect if file exists
	var _, err = os.Stat(path)
	if err == nil {
		return true, err
	}
	// create dirs if file not exists
	if os.IsNotExist(err) {
		if filepath.Dir(path) != "." {
			return false, os.MkdirAll(filepath.Dir(path), 0777)
		}
	}
	return false, err
}
