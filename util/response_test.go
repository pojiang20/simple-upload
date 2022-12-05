package util

import (
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_writeError(t *testing.T) {
	r := assert.New(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/testWriteError", func(writer http.ResponseWriter, request *http.Request) {
		writeError(writer, errors.New("testError"))
	})
	reader := strings.NewReader("test content")
	req, err := http.NewRequest(http.MethodPost, "/testWriteError", reader)
	r.NoError(err)

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	resp := w.Result()
	data, err := io.ReadAll(resp.Body)
	r.NoError(err)

	ret := errorResponse{}
	err = json.Unmarshal(data, &ret)
	r.NoError(err)
	t.Log(ret)

	r.Equal(errorResponse{response{false}, "testError"}, ret)
}

func Test_writeSuccess(t *testing.T) {
	r := assert.New(t)
	testPath := "/test/success"

	mux := http.NewServeMux()
	mux.HandleFunc("/testWriteSuccess", func(writer http.ResponseWriter, request *http.Request) {
		writeSuccess(writer, testPath)
	})
	reader := strings.NewReader("test content")
	req, err := http.NewRequest(http.MethodPost, "/testWriteSuccess", reader)
	r.NoError(err)

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	resp := w.Result()
	data, err := io.ReadAll(resp.Body)
	r.NoError(err)

	ret := uploadedResponse{}
	err = json.Unmarshal(data, &ret)
	r.NoError(err)
	t.Log(ret)

	r.Equal(uploadedResponse{response{true}, testPath}, ret)
}
