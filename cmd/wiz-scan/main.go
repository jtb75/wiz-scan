package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/jtb75/wiz-scan/pkg/utilities"
	"github.com/jtb75/wiz-scan/pkg/wizapi"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

// Set testRun to "false", "genData", or "runTestData"
var runTest = "runTestData"

func LogInit(level string) {
	// Map string log level to logrus.Level
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		log.Fatalf("Invalid log level: %s", level)
	}

	log.SetLevel(logLevel)
}

func gatherWizKnownVulns(runTest string, wizAPI *wizapi.WizAPI, resourceID string) (interface{}, error) {
	var response interface{}
	var err error

	if runTest == "genData" || runTest == "runTestData" {
		if runTest == "genData" {
			response, err = wizapi.FetchAllVulnerabilities(wizAPI, resourceID)
			if err != nil {
				return nil, fmt.Errorf("error fetching vulnerabilities: %v", err)
			}

			jsonResponseBytes, err := json.MarshalIndent(response, "", "    ")
			if err != nil {
				return nil, fmt.Errorf("error marshalling JSON: %v", err)
			}

			err = os.WriteFile("sample_data/known_vulns.json", jsonResponseBytes, 0644)
			if err != nil {
				return nil, fmt.Errorf("error writing to file: %v", err)
			}
		} else {
			data, err := os.ReadFile("sample_data/known_vulns.json")
			if err != nil {
				return nil, fmt.Errorf("failed to read sample data file: %v", err)
			}

			if err := json.Unmarshal(data, &response); err != nil {
				return nil, fmt.Errorf("failed to unmarshal sample data: %v", err)
			}
		}
	} else {
		response, err = wizapi.FetchAllVulnerabilities(wizAPI, resourceID)
		if err != nil {
			return nil, fmt.Errorf("error fetching vulnerabilities: %v", err)
		}
	}

	return response, nil
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

	response, err := gatherWizKnownVulns(runTest, wizAPI, resourceID)
	if err != nil {
		log.Errorf("Error gathering known vulnerabilities: %v", err)
		return
	}

	// Marshal the response variable into a JSON string with indentation
	jsonResponse, err := json.MarshalIndent(response, "", "    ")
	if err != nil {
		log.Errorf("Error marshalling JSON response: %v", err)
		return
	}

	// Print the JSON string
	fmt.Println(string(jsonResponse))
}
