package util

import "errors"

var (
	ErrInvalidArgument = errors.New("invalid argument")
	ErrNotExistKey     = errors.New("key dose not exist")
)
