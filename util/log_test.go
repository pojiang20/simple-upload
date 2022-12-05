package util

import "testing"

func Test_Log(t *testing.T) {
	Zlog.Info("test info")
	Zlog.Error("test error")
}
