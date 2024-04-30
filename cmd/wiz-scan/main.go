package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/jtb75/wiz-scan/pkg/utilities"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func LogInit(level logrus.Level) {
	log.SetLevel(level)
}

func main() {

	// Initialize logging with default Info level
	LogInit(logrus.InfoLevel)

	// Get the detected operating system
	operatingSystem := runtime.GOOS

	// Print the detected operating system
	log.Debug("Operating System:", operatingSystem)

	args, err := utilities.ProcessArguments() // Capture both the arguments and the error
	if err != nil {
		// Log the error and exit if ArgParse encountered an issue
		log.Errorf("Failed to parse arguments: %v", err)
		os.Exit(1) // Exit the program with a non-zero status indicating failure
	}

	// If uninstall flag is passed, initiate process
	if args.Uninstall {
		log.Info("Initiating Uninstall")
		os.Exit(0)
	}

	// If install flag is passed, initiate process
	if args.Install {
		log.Info("Initiating Install")
		os.Exit(0)
	}
	fmt.Println(args)
}
