package util

import (
	"encoding/json"
	"net/http"
)

type response struct {
	OK bool `json:"ok"`
}

type uploadedResponse struct {
	response
	Path string `json:"path"`
}

func newUploadedResponse(path string) uploadedResponse {
	return uploadedResponse{
		response: response{
			true,
		},
		Path: path,
	}
}

type errorResponse struct {
	response
	Message string `json:"message"`
}

func newErrorResponse(err error) errorResponse {
	return errorResponse{
		response: response{false},
		Message:  err.Error(),
	}
}

func writeError(w http.ResponseWriter, wErr error) (int, error) {
	body := newErrorResponse(wErr)
	b, err := json.Marshal(body)
	if err != nil {
		return w.Write([]byte{})
	}
	return w.Write(b)
}

func writeSuccess(w http.ResponseWriter, path string) (int, error) {
	body := newUploadedResponse(path)
	b, err := json.Marshal(body)
	if err != nil {
		return w.Write([]byte{})
	}
	return w.Write(b)
}
