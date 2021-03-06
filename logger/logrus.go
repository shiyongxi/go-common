package logger

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

type (
	Logger struct {
		Debug         bool     `yaml:"debug"`
		StdOut        string   `yaml:"stdOut"`
		AppName       string   `yaml:"appName"`
		FileName      string   `yaml:"fileName"`
		SavePath      string   `yaml:"savePath"`
		LogStashHost  string   `yaml:"logStashHost"`
		LogStashPort  int      `yaml:"logStashPort"`
		RedisHost     string   `yaml:"redisHost"`
		RedisPort     int      `yaml:"redisPort"`
		RedisDB       int      `yaml:"redisDB"`
		RedisKey      string   `yaml:"redisKey"`
		RedisPassword string   `yaml:"redisPassword"`
		Brokers       []string `yaml:"brokers"`
		Topics        string   `yaml:"topics"`
		ElasticHost   string   `yaml:"elasticHost"`
		ElasticPost   int      `yaml:"elasticPost"`
		PrefixIndex   string   `yaml:"prefixIndex"`
	}
)

var (
	LLogger *logrus.Logger
)

func NewLogger(logger *Logger) {
	stdOut := strings.ToLower(logger.StdOut)
	switch stdOut {
	case "logstash":
		logger := NewLogStash(logger.AppName, logger.LogStashHost, logger.LogStashPort).Output()
		LLogger = logger.Logger
		break
	case "redis":
		logger := NewRedis(logger.AppName, logger.RedisHost, logger.RedisKey, logger.RedisPassword, logger.RedisDB, logger.RedisPort).Output()
		LLogger = logger.Logger
		break
	case "elasticsearch":
		logger := NewElastic(logger.ElasticHost, logger.PrefixIndex, logger.ElasticPost).Output()
		LLogger = logger.Logger
		break
	default:
		logger := NewFile(logger.SavePath, logger.FileName, logger.Debug).Output()
		LLogger = logger.Logger
		break
	}

	if logger.Debug {
		LLogger.SetLevel(logrus.DebugLevel)
	}
}

func GetLogger() *logrus.Logger {
	if LLogger == nil {
		LLogger = logrus.New()
	}

	return LLogger
}

func Info(message ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	files := fmt.Sprintf("%s (%d)", file, line)

	GetLogger().WithFields(logrus.Fields{
		"files": files,
	}).Info(message)
}

func Infos(msg string, message ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	files := fmt.Sprintf("%s (%d)", file, line)

	body, err := json.Marshal(message)
	if err != nil {
		Error(err)
	}

	GetLogger().WithFields(logrus.Fields{
		"files":   files,
		"message": string(body),
	}).Info(msg)
}

func Warn(err error, message ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	files := fmt.Sprintf("%s (%d)", file, line)

	GetLogger().WithFields(logrus.Fields{
		"files":  files,
		"errors": err,
	}).Warn(message)
}

func Fatal(err error, message ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	files := fmt.Sprintf("%s (%d)", file, line)

	GetLogger().WithFields(logrus.Fields{
		"files":  files,
		"errors": err,
	}).Fatal(message)
}

func Error(err error, message ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	files := fmt.Sprintf("%s (%d)", file, line)

	GetLogger().WithFields(logrus.Fields{
		"files":  files,
		"errors": err,
	}).Error(message)

}

func Debug(message ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	files := fmt.Sprintf("%s (%d)", file, line)

	GetLogger().WithFields(logrus.Fields{
		"files": files,
	}).Debug(message)
}
