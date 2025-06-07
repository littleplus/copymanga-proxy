package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	listen    string
	tlsCert   string
	tlsKey    string
	debugMode bool

	trustProxy = []string{"10.0.0.0/8", "192.168.0.0/16", "172.16.0.0/12", "127.0.0.0/8"}
)

func init() {
	flag.StringVar(&listen, "l", ":8080", "listen host and port")
	flag.StringVar(&tlsCert, "c", "", "enable tls(debug):tls cert")
	flag.StringVar(&tlsKey, "k", "", "enable tls(debug):tls key")
	flag.BoolVar(&debugMode, "d", false, "enable debug http response print")
}

func proxyHandler(c *gin.Context) {
	remote, err := url.Parse("https://api.mangacopy.com")
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.ModifyResponse = func(r *http.Response) error {
		resp, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewReader(resp))
		gzReader, err := gzip.NewReader(bytes.NewReader(resp))
		var gzipFlag = false
		if err != nil {
			fmt.Println("decompress gzip error: ", err)
		} else {
			resp, _ = io.ReadAll(gzReader)
			gzipFlag = true
		}

		respString := string(resp)
		if debugMode {
			fmt.Printf("[debug] -> [%v]: \nHeader: %v\nResp: %v\n", r.Request.URL, r.Request.Header, respString)
		}
		rega := regexp.MustCompile(`([a-zA-Z0-9\-]+)\.mangafuna\.xyz`)
		regb := regexp.MustCompile(`([a-zA-Z0-9\-]+)\.mangafunb\.fun`)
		respString = rega.ReplaceAllString(respString, c.Request.Host+`/a/$1`)
		respString = regb.ReplaceAllString(respString, c.Request.Host+`/b/$1`)
		respString = strings.ReplaceAll(respString, "api.mangacopy.com", c.Request.Host)
		respString = strings.ReplaceAll(respString, "t66y.com", c.Request.Host)
		// fmt.Println("data: ", respString)
		// fmt.Println("header: ", r.Header)
		if debugMode {
			fmt.Printf("RespReplaced: %v\n", respString)
		}
		resp = []byte(respString)

		if gzipFlag {
			var respBuff bytes.Buffer
			gzWriter := gzip.NewWriter(&respBuff)
			gzWriter.Write([]byte(resp))
			gzWriter.Close()
			resp = respBuff.Bytes()
			r.ContentLength = int64(len(resp))
			r.Header.Set("content-length", strconv.Itoa(int(r.ContentLength)))
			r.Body = io.NopCloser(bytes.NewReader(resp))
		} else {
			r.Body = io.NopCloser(bytes.NewReader(resp))
		}
		return nil
	}
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = c.Param("anypath")
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

func main() {
	flag.Parse()

	r := gin.Default()
	r.SetTrustedProxies(trustProxy)
	r.Any("/*anypath", proxyHandler)
	if tlsCert != "" && tlsKey != "" {
		r.RunTLS(":443", tlsCert, tlsKey)
	} else {
		r.Run(listen)
	}
}
