package main

import (
	"os"
	"runtime"

	"github.com/jtb75/wiz-scan/pkg/utilities"
	"github.com/jtb75/wiz-scan/pkg/wizapi"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func LogInit(level string) {
	// Map string log level to logrus.Level
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		log.Fatalf("Invalid log level: %s", level)
	}

	log.SetLevel(logLevel)
}

func main() {

	// Initialize logging with default Info level
	LogInit("info") // Set default log level to Info

	// Get the detected operating system
	operatingSystem := runtime.GOOS

	args, err := utilities.ProcessArguments() // Capture both the arguments and the error
	if err != nil {
		// Log the error and exit if ArgParse encountered an issue
		log.Errorf("Failed to parse arguments: %v", err)
		os.Exit(1) // Exit the program with a non-zero status indicating failure
	}

	// Set log level based on arguments
	LogInit(args.LogLevel)

	// Print the detected operating system
	log.Debug("Operating System:", operatingSystem)

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

	// Create a new instance of WizAPI
	wizAPI, err := wizapi.NewWizAPI(
		args.WizClientID,
		args.WizClientSecret,
		args.WizAuthURL,
		args.WizQueryURL,
	)
	if err != nil {
		log.Errorf("Failed to create WizAPI instance: %v", err)
		os.Exit(1)
	}
	log.Infof("Auth Token: %s", wizAPI.AuthToken)

	// Call GetResourceID method using args.ScanCloudType and args.ScanProviderID
	resourceID, err := wizAPI.GetResourceID(args.ScanCloudType, args.ScanProviderID)
	if err != nil {
		log.Errorf("Failed to get resource ID: %v", err)
		os.Exit(1)
	}

	log.Debugf("Matched Resource ID: %s", resourceID)

}
