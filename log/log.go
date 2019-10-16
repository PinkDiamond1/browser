package log

import (
	"fmt"
	"github.com/browser_service/config"
	"github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"time"
)

const (
	defaultLevel          = "debug"
	defaultTimeKey        = "ts"
	defaultLevelKey       = "level"
	defaultMessageKey     = "msg"
	defaultCallerKey      = "caller"
	dayLogFileNameSuffix  = ".%Y%m%d"
	hourLogFileNameSuffix = ".%Y%m%d%H"
)

var ZapLog *zap.Logger

func InitLog() {
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.TimeKey = defaultTimeKey
	encoderConfig.LevelKey = defaultLevelKey
	encoderConfig.MessageKey = defaultMessageKey
	encoderConfig.CallerKey = defaultCallerKey
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	level := getLogLevel()
	cores := make([]zapcore.Core, 0)
	if config.Log.FileConfig.Enable {
		logFileWriter := getWriter(config.Log.FileConfig.Path)
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(logFileWriter), level))
	}

	if config.Log.Console {
		cores = append(cores, zapcore.NewCore(encoder, os.Stdout, level))
	}

	core := zapcore.NewTee(cores...)
	ZapLog = zap.New(core, zap.AddCaller())
	ZapLog.Info("ZapLogger construction succeeded")
}

func getLogLevel() zap.AtomicLevel {
	logLevel := config.Log.Level
	if logLevel == "" {
		logLevel = defaultLevel
	}

	level := zap.NewAtomicLevel()
	err := level.UnmarshalText([]byte(logLevel))
	if err != nil {
		panic(fmt.Sprint("log level configuration error. unsupported log level: ", logLevel))
	}
	return level
}

func getWriter(fileName string) io.Writer {
	nameSuffix, rotationTime := getLogRotationTime()
	writer, err := rotatelogs.New(
		fileName+nameSuffix,
		rotatelogs.WithLinkName(fileName),
		rotatelogs.WithMaxAge(time.Hour*24*time.Duration(config.Log.FileConfig.MaxAge)),
		rotatelogs.WithRotationTime(rotationTime),
	)

	if err != nil {
		panic(err)
	}

	return writer
}

func getLogRotationTime() (string, time.Duration) {
	rotationTime := config.Log.FileConfig.RotationTime
	if rotationTime == 2 {
		return hourLogFileNameSuffix, time.Hour
	} else {
		return dayLogFileNameSuffix, time.Hour * 24
	}
}
