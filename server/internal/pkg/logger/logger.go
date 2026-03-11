package logger

import (
	"log"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *zap.Logger

func Init() {
	logDir := "logs"
	logFile := filepath.Join(logDir, "app.log")

	err := os.MkdirAll(logDir, 0755)

	if err != nil {
		log.Fatalf("Не удалось создать деректорию для логов, %v", err)
	}

	//Настройка для долговечности логов
	//коротко, что бы логи не весили по несколько гигов спустя время
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     14,
		Compress:   true,
	})

	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncodeLevel = zapcore.CapitalLevelEncoder

	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewConsoleEncoder(config), w, zap.InfoLevel),
		zapcore.NewCore(zapcore.NewJSONEncoder(config), zapcore.AddSync(os.Stdout), zap.InfoLevel),
	)

	Log = zap.New(core, zap.AddCaller())
}
