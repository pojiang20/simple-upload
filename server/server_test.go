package server

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"os"
	"os/exec"
	"testing"
)

func TestServer_Run(t *testing.T) {
	r := assert.New(t)
	server := NewServer(1024, "aa")

	address := "127.0.0.1:1234"
	go server.Run(address)
	resp, err := http.Get(fmt.Sprintf("http://%s/zenfs", address))
	r.NoError(err)

	data, err := io.ReadAll(resp.Body)
	r.NoError(err)
	r.Contains(string(data), "Welcome")
}

func TestUploadAndGetFile(t *testing.T) {
	r := assert.New(t)
	err := os.RemoveAll("./public")
	r.NoError(err)

	server := NewServer(1024, "aa")
	r.NotNil(server)

	address := "127.0.0.1:1234"
	go server.Run(address)

	url := fmt.Sprintf("http://%s/zenfs/upload", address)
	curl := exec.Command("curl", "-Ffile=@./testfile.txt", url)
	out, err := curl.Output()
	r.NoError(err)
	t.Log(string(out))

	resp, err := http.Get(fmt.Sprintf("http://%s/zenfs/public/testfile.txt", address))
	r.NoError(err)
	data, err := io.ReadAll(resp.Body)
	r.NoError(err)
	t.Log(string(data))
}
