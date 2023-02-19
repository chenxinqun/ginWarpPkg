package loggerx

import "go.uber.org/zap"

func CreateSpecialLogger() *zap.Logger {
	cfg := new(LoggersInfo)
	l, _ := New(cfg)
	return l
}
