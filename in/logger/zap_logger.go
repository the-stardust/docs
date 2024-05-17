package logger

import (
	"interview/common/global"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitZapLogger() {
	logConfig := global.CONFIG.Log
	encoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		TimeKey:        "ts",
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
		CallerKey:      "caller",
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.NanosDurationEncoder,
	})

	// 实现两个判断日志等级的interface
	debugLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.DebugLevel
	})
	proLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.InfoLevel
	})
	tee := []zapcore.Core{}
	if logConfig.Debug {
		tee = append(tee, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), debugLevel))
	} else {
		tee = append(tee, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), proLevel))
	}
	// 最后创建具体的Logger
	core := zapcore.NewTee(
		tee...,
	)
	logger := zap.New(core, zap.AddCaller())
	// logger = logger.With(zap.String("svc", "interview"))
	global.LOGGER = logger
	global.SUGARLOGGER = global.LOGGER.Sugar()
}
