package util

import "errors"

var (
	ErrInvalidArgument = errors.New("invalid argument")
	ErrKeyNotExist     = errors.New("key dose not exist")
)
