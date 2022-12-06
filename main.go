package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"github.com/pojiang20/simple-upload/util"
	"net/http"
	"os"
)

const (
	DEFAULTPORT = 23456
	MAXFILESIZE = 1 << 22
)

func main() {
	util.Zlog.Info("starting server")
	run(os.Args)
}

func run(args []string) {
	address := flag.String("ip", "127.0.0.1", "IP address")
	port := flag.Int("port", DEFAULTPORT, "listen port")
	maxUploadSize := flag.Int64("upload_limit", MAXFILESIZE, "max file byte size")
	tokenFlag := flag.String("token", "", "specify the security token (it is automatically generated if empty)")
	rootPath := flag.String("filePath", "/tmp/simpleUpload", "file system root path")
	flag.Parse()

	token := *tokenFlag
	if token == "" {
		b := make([]byte, 10)
		_, err := rand.Read(b)
		if err != nil {
			return
		}
		token = fmt.Sprintf("%x", b)
	}
	util.Zlog.Infof("ip %s,port %d,token %s,upload_limit %d,root %s", *address, *port, token, *maxUploadSize, *rootPath)
	server := NewServer(*rootPath, *maxUploadSize, token)
	http.Handle("/upload", server)
	http.Handle("/files/", server)
	http.ListenAndServe(fmt.Sprintf("%s:%d", *address, *port), nil)
	return
}
