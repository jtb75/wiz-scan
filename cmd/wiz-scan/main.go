package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/jtb75/wiz-scan/pkg/utilities"
	"github.com/jtb75/wiz-scan/pkg/vulnerability"
	"github.com/jtb75/wiz-scan/pkg/wizapi"
	"github.com/jtb75/wiz-scan/pkg/wizcli"
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

func scanDirectories(drives []string, aggregatedResults *wizcli.AggregatedScanResults, operatingSystem string, wizCliPath string) error {
	if runTest != "runTestData" {
		for _, drive := range drives {
			mountedPath := ""
			shadowCopyID := ""
			// If Windows, initiate VSS snapshot
			if operatingSystem == "windows" {
				var err error // Define err here
				mountedPath, shadowCopyID, err = utilities.CreateVSSSnapshot(drive)
				if err != nil {
					log.Errorf("Error creating VSS snapshot for drive %s: %v", drive, err)
					if err := RemoveSymbolicLink(mountedPath); err != nil {
						log.Errorf("Failed to remove symbolic link: %v", err)
					}
					continue
				} else {
					log.Infof("Created VSS ID `%s` Mounted on: %s", shadowCopyID, mountedPath)
				}
			}
			if mountedPath == "" {
				mountedPath = drive
			}
			scanResult, err := wizcli.ScanDirectory(wizCliPath, mountedPath)
			if err != nil {
				log.Errorf("Failed to scan %s: %v", mountedPath, err)
				continue
			} else {
				log.Info("Scanned successfully")
			}
			// Prepend the Drive to the Library path to represent actual full path
			for i, lib := range scanResult.Result.Libraries {
				if runtime.GOOS == "windows" {
					lib.Path = strings.ReplaceAll(lib.Path, "/", "\\")
					lib.Path = strings.TrimPrefix(lib.Path, "\\")
				}
				scanResult.Result.Libraries[i].Path = drive + lib.Path
			}
			aggregatedResults.Libraries = append(aggregatedResults.Libraries, scanResult.Result.Libraries...)
			aggregatedResults.Applications = append(aggregatedResults.Applications, scanResult.Result.Applications...)
			// If Windows, clean up VSS snapshot
			if operatingSystem == "windows" {
				if err := utilities.RemoveVSSSnapshot(mountedPath, shadowCopyID); err != nil {
					log.Errorf("Failed to remove mount and VSS snapshot for drive %s: %v", drive, err)
				} else {
					log.Infof("Removed mount and VSS snapshot for drive %s", drive)
				}
			}
		}
	}
	if runTest == "genData" {
		jsonBytes, err := json.MarshalIndent(aggregatedResults, "", "    ")
		if err != nil {
			log.Errorln("Error marshalling JSON:", err)
		}

		err = os.WriteFile("sample_data/scan.json", jsonBytes, 0644)
		if err != nil {
			log.Errorln("Error writing to file:", err)
		}
		os.Exit(0)
	}
	if runTest == "runTestData" {
		// Read in sample data into aggregatedResults
		log.Infoln("Reading in sample scan data")
		data, err := os.ReadFile("sample_data/scan.json")
		if err != nil {
			log.Errorln("Error reading sample data file:", err)
			return err
		}
		if err := json.Unmarshal(data, &aggregatedResults); err != nil {
			log.Errorln("Error unmarshalling JSON:", err)
			return err
		}
	}

	return nil
}

func gatherWizKnownVulns(runTest string, wizAPI *wizapi.WizAPI, resourceID string) ([]wizapi.VulnerabilityNode, error) {
	var response []wizapi.VulnerabilityNode
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
			log.Infoln("Reading in sample known data")
			data, err := os.ReadFile("sample_data/known_vulns.json")
			if err != nil {
				return nil, fmt.Errorf("failed to read sample data file: %v", err)
			}

			if err := json.Unmarshal(data, &response); err != nil {
				return nil, fmt.Errorf("failed to unmarshal sample data: %v", err)
			}
		}
	} else {
		// Fetch vulnerabilities directly if not in test mode
		response, err = wizapi.FetchAllVulnerabilities(wizAPI, resourceID)
		if err != nil {
			return nil, fmt.Errorf("error fetching vulnerabilities: %v", err)
		}
	}

	return response, nil
}

func RemoveSymbolicLink(path string) error {
	// RemoveSymbolicLink removes the symbolic link created by CreateVSSSnapshot
	err := os.Remove(path)
	if err != nil {
		return fmt.Errorf("failed to remove symbolic link: %v", err)
	}
	return nil
}

func main() {

	var assetVulns vulnerability.Asset

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
	fmt.Print(args.Install)
	// Print the detected operating system
	log.Debug("Operating System:", operatingSystem)

	// If uninstall flag is passed, initiate process
	if args.Uninstall {
		log.Info("Initiating Uninstall")
		err := utilities.UninstallAndRemoveTask()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("Uninstallation and task removal completed successfully.")
		os.Exit(0)
	}

	// If install flag is passed, initiate process
	if args.Install {
		log.Info("Initiating Install")
		err := utilities.InstallAndScheduleTask()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("Installation and task scheduling completed successfully.")
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
	log.Debugf("Auth Token: %s", wizAPI.AuthToken)

	// Call GetResourceID method using args.ScanCloudType and args.ScanProviderID
	resourceID, err := wizAPI.GetResourceID(args.ScanCloudType, args.ScanProviderID)
	if err != nil {
		log.Errorf("Failed to get resource ID: %v", err)
		os.Exit(1)
	}
	log.Debugf("Matched Resource ID: %s", resourceID)

	log.Info("Gathering known vulnerabilities from Wiz platform")
	response, err := gatherWizKnownVulns(runTest, wizAPI, resourceID)
	if err != nil {
		log.Errorf("Error gathering known vulnerabilities: %v", err)
		return
	}

	// Initialize and authenticate wizcli
	cleanup, wizCliPath, err := wizcli.InitializeAndAuthenticate(args.WizClientID, args.WizClientSecret)
	if err != nil {
		log.Errorf("initialization and authentication failed: %v", err)
		return
	}
	defer cleanup()

	// Retrieve top-level directories
	directories, err := utilities.GetTopLevelDirectories()
	if err != nil {
		log.Errorf("Error listing directories: %v", err)
		return
	} else {
		log.Debug("Directories to scan: ", directories)
	}

	aggregatedResults := wizcli.AggregatedScanResults{}

	log.Info("Initiating directory scan")
	// Cycle through directories and initiate scan
	//directories = []string{"E:\\"}
	if err := scanDirectories(directories, &aggregatedResults, operatingSystem, wizCliPath); err != nil {
		log.Errorf("Error scanning directories: %v", err)
		return
	}

	assetVulns, err = vulnerability.CompareVulnerabilities(aggregatedResults, response, args.ScanProviderID)
	if err != nil {
		fmt.Printf("Error in CompareVulnerabilities: %s\n", err)
		return
	}

	var vulnPayloadJSON []byte // Use a byte slice to hold JSON data

	if len(assetVulns.VulnerabilityFindings) > 0 {
		assetVulns.AssetIdentifier.CloudPlatform = args.ScanCloudType
		assetVulns.AssetIdentifier.ProviderId = args.ScanProviderID
		vulnPayload := vulnerability.IntegrationData{
			IntegrationId: "e4341955-463f-4228-aa99-a718e9d93bb5", // Set an integration ID
			DataSources:   []vulnerability.DataSource{},           // Initialize an empty slice of DataSources
		}
		// Create a DataSource and add assetVulns to it
		dataSource := vulnerability.DataSource{
			Id:           args.ScanSubscriptionID,
			AnalysisDate: time.Now(),                        // Set current time as the analysis date
			Assets:       []vulnerability.Asset{assetVulns}, // Add assetVulns here
		}

		vulnPayload.DataSources = append(vulnPayload.DataSources, dataSource)

		vulnPayloadJSON, err = json.MarshalIndent(vulnPayload, "", "\t")
		if err != nil {
			fmt.Println("Error marshaling assetVulns to JSON:", err)
			return
		}
	} else {
		log.Infof("No new vulnerabilities found")
		return // Exit the program gracefully
	}

	file, err := utilities.CreateTempFile()
	if err != nil {
		log.Errorln("Error creating temp file:", err)
		return
	}

	defer func() {
		// Ensure the temporary file is deleted upon exiting the function
		if err := file.Close(); err != nil {
			log.Errorf("Error closing file: %v", err)
		}
		if err := os.Remove(file.Name()); err != nil {
			log.Errorf("Error removing temporary file: %v", err)
		}
	}()

	log.Infoln("Temporary file created:", file.Name())
	_, err = file.Write(vulnPayloadJSON)

	if err != nil {
		log.Errorln("Error writing JSON to temp file:", err)
		return
	}

	if err := wizAPI.PublishVulns(file.Name()); err != nil {
		log.Errorln("Error publishing vulnerabilities:", err)
	}

}
