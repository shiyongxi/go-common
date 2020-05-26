package logger

import (
	"fmt"
	"github.com/bshuster-repo/logrus-logstash-hook"
	"github.com/sirupsen/logrus"
)

type (
	LogStashConfig struct {
		Logger       *logrus.Logger
		LogStashHost string
		LogStaShPort int
		AppName      string
	}
)

func NewLogStash(appName, logStashHost string, logStaShPort int) *LogStashConfig {
	return &LogStashConfig{
		Logger:       logrus.New(),
		AppName:      appName,
		LogStashHost: logStashHost,
		LogStaShPort: logStaShPort,
	}
}

func (log *LogStashConfig) Output() *LogStashConfig {
	hook, err := logrustash.NewHook("tcp", log.getLogStashAddress(), log.AppName)
	if err != nil {
		log.Logger.Fatal(err)
	}

	log.Logger.Formatter = &logrus.JSONFormatter{}
	log.Logger.Level = logrus.InfoLevel
	log.Logger.Hooks.Add(hook)

	return log
}

func (log *LogStashConfig) getLogStashAddress() string {
	return fmt.Sprintf("%s:%d", log.LogStashHost, log.LogStaShPort)
}
