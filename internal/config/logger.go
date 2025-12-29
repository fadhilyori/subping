// Package config provides global configuration for the subping application
package config

import (
	"github.com/sirupsen/logrus"
	"os"
)

func init() {
	level := os.Getenv("SUBPING_LOG_LEVEL")
	if level == "" {
		level = "error" // Default level
	}

	if parsedLevel, err := logrus.ParseLevel(level); err == nil {
		logrus.SetLevel(parsedLevel)
	} else {
		logrus.SetLevel(logrus.ErrorLevel) // Fallback to error
	}
}

// GetLogLevel returns the current global log level
func GetLogLevel() logrus.Level {
	return logrus.GetLevel()
}
