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

func WriteError(w http.ResponseWriter, statusCode int, wErr error) (int, error) {
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}
	w.WriteHeader(statusCode)
	body := newErrorResponse(wErr)
	b, err := json.Marshal(body)
	if err != nil {
		Zlog.Info("writeError Marshal error")
		return w.Write([]byte{})
	}
	return w.Write(b)
}

func WriteSuccess(w http.ResponseWriter, path string) (int, error) {
	w.WriteHeader(http.StatusOK)
	body := newUploadedResponse(path)
	b, err := json.Marshal(body)
	if err != nil {
		Zlog.Info("writeSuccess Marshal error")
		return w.Write([]byte{})
	}
	return w.Write(b)
}
