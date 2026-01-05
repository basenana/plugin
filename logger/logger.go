package logger

import "go.uber.org/zap"

var (
	root *zap.SugaredLogger
)

func SetLogger(log *zap.SugaredLogger) {
	root = log
}

func NewLogger(name string) *zap.SugaredLogger {
	return root.Named(name)
}
