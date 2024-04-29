package main

import (
	"runtime"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func LogInit(level logrus.Level) {
	log.SetLevel(level)
}

func main() {

	// Initialize logging
	LogInit(logrus.InfoLevel)

	// Get the detected operating system
	operatingSystem := runtime.GOOS

	// Print the detected operating system
	log.Info("Operating System:", operatingSystem)
}
