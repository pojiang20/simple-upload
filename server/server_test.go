package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"
)

const MONGODB_ADDR = "127.0.0.1:27017"

func TestServer_Run(t *testing.T) {
	r := assert.New(t)
	dbAddr := MONGODB_ADDR
	server := NewServer(1024, "aa", dbAddr)

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

	dbAddr := MONGODB_ADDR
	server := NewServer(1024, "aa", dbAddr)
	r.NotNil(server)

	address := "127.0.0.1:1234"
	go server.Run(address)

	url := fmt.Sprintf("http://%s/zenfs/upload", address)
	body := strings.NewReader("test data to upload")
	resp, err := http.Post(url, "application/json; charset=utf-8", body)
	r.NoError(err)
	t.Log(resp.Body)

	resp1, err := http.Get(fmt.Sprintf("http://%s/zenfs/public/testfile.txt", address))
	r.NoError(err)
	data, err := io.ReadAll(resp1.Body)
	r.NoError(err)
	t.Log(string(data))
}

func TestServer_partUpload(t *testing.T) {
	r := assert.New(t)
	err := os.RemoveAll("./public")
	r.NoError(err)

	dbAddr := MONGODB_ADDR
	key := "file1"
	server := NewServer(1024, "aa", dbAddr)
	r.NotNil(server)

	address := "127.0.0.1:1234"
	go server.Run(address)

	dbHelp := newMongoDB(MONGODB_ADDR)
	err = dbHelp.Database(DATABASE_PART).Drop(context.TODO())
	r.NoError(err)

	url := fmt.Sprintf("http://%s/zenfs/uploads/init?key=%s", address, key)
	body := strings.NewReader("")
	resp, err := http.Post(url, "application/json; charset=utf-8", body)
	r.NoError(err)
	data, err := io.ReadAll(resp.Body)
	r.NoError(err)
	res := respMsg{}
	err = json.Unmarshal(data, &res)
	r.NoError(err)
	r.Contains(res.Msg, "success")
	uploadId := res.Info.(map[string]interface{})["upload_id"]
	t.Log(uploadId)

	content1 := "aaa"
	url1 := fmt.Sprintf("http://%s/zenfs/uploads/uploadPart?UploadId=%s&key=%s&PartNumber=1", address, uploadId, key)
	body1 := strings.NewReader(content1)
	resp1, err := http.Post(url1, "application/json; charset=utf-8", body1)
	r.NoError(err)
	data, err = io.ReadAll(resp1.Body)
	r.NoError(err)
	res = respMsg{}
	err = json.Unmarshal(data, &res)
	r.NoError(err)
	r.Contains(res.Msg, "success")
	etags1 := res.Info.(map[string]interface{})["etags"]
	t.Log(etags1)

	content2 := "bbb"
	url2 := fmt.Sprintf("http://%s/zenfs/uploads/uploadPart?UploadId=%s&key=%s&PartNumber=2", address, uploadId, key)
	body2 := strings.NewReader(content2)
	resp2, err := http.Post(url2, "application/json; charset=utf-8", body2)
	r.NoError(err)
	data, err = io.ReadAll(resp2.Body)
	r.NoError(err)
	res = respMsg{}
	err = json.Unmarshal(data, &res)
	r.NoError(err)
	r.Contains(res.Msg, "success")
	etags2 := res.Info.(map[string]interface{})["etags"]
	t.Log(etags2)

	url3 := fmt.Sprintf("http://%s/zenfs/uploads/complete?uploadId=%s&key=%s", address, uploadId, key)
	body3 := strings.NewReader("[{\"Etag\":\"c1\",\"PartNumber\":1,\"PartSize\":193},{\"Etag\":\"bd\",\"PartNumber\":2,\"PartSize\":189}]\n")
	resp3, err := http.Post(url3, "application/json; charset=utf-8", body3)
	r.NoError(err)
	t.Log(resp3.Body)

	data, err = os.ReadFile(path.Join("./public", key))
	r.NoError(err)
	r.Equal(content1+content2, string(data))
	t.Log(string(data))
}
