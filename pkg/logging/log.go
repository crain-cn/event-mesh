package logging

import (
	"bufio"
	"github.com/lestrrat/go-file-rotatelogs"
	"github.com/pkg/errors"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
	"time"
)

var (
	DefaultLogger = logrus.New()
	log           = DefaultLogger.WithField("component", "Log")
)

func InitLogger() *logrus.Logger {
	baseLogPath := path.Join("/data/eventmesh/logs", "api.log")

	fileWriter, err := rotatelogs.New(
		baseLogPath+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(baseLogPath),      // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(7*24*time.Hour),     // 文件最大保存时间
		rotatelogs.WithRotationTime(24*time.Hour), // 日志切割时间间隔
	)

	stdWriter := os.Stderr
	writer := io.MultiWriter(fileWriter, stdWriter)

	if err != nil {
		DefaultLogger.Errorf("config local file system logger error. %v", errors.WithStack(err))
	}

	//log.SetFormatter(&log.TextFormatter{})
	switch level := "info"; level {
	case "debug":
		DefaultLogger.SetLevel(logrus.DebugLevel)
		DefaultLogger.SetOutput(os.Stderr)
	case "info":
		setNull()
		DefaultLogger.SetLevel(logrus.InfoLevel)
	case "warn":
		setNull()
		DefaultLogger.SetLevel(logrus.WarnLevel)
	case "error":
		setNull()
		DefaultLogger.SetLevel(logrus.ErrorLevel)
	default:
		setNull()
		DefaultLogger.SetLevel(logrus.InfoLevel)
	}

	lfHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: writer,
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
		logrus.PanicLevel: writer,
	}, &logrus.JSONFormatter{
		PrettyPrint:       false,
		DisableHTMLEscape: true,
		TimestampFormat:   "2006-01-02 15:04:05",
	})
	DefaultLogger.AddHook(lfHook)
	log.Info("InitLogger...")
	return DefaultLogger
}

func setNull() {
	src, err := os.OpenFile(os.DevNull, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		DefaultLogger.Errorf("open file DevNull. %v", errors.WithStack(err))
	}
	writer := bufio.NewWriter(src)
	DefaultLogger.SetOutput(writer)
}
