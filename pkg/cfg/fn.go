package cfg

import "github.com/sirupsen/logrus"

var logger = logrus.New()

func init() {
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.999-0700",
	})
}

func LoggerForTask(task string) *logrus.Entry {
	return logger.WithField("task", task)
}
