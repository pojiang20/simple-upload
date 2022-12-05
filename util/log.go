package util

import "go.uber.org/zap"

var Zlog *zap.SugaredLogger

func init() {
	l, _ := zap.NewProduction(zap.AddCallerSkip(1))
	Zlog = l.Sugar()
}
