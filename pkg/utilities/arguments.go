package utilities

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type Arguments struct {
	WizClientID        string `json:"wizClientId"`
	WizClientSecret    string `json:"wizClientSecret"`
	WizQueryURL        string `json:"wizQueryUrl"`
	WizAuthURL         string `json:"wizAuthUrl"`
	ScanSubscriptionID string `json:"scanSubscriptionId"`
	ScanCloudType      string `json:"scanCloudType"`
	ScanProviderID     string `json:"scanProviderId"`
	Save               bool   `json:"save"`
	Install            bool   `json:"install"`
	Uninstall          bool   `json:"uninstall"`
	LogLevel           string `json:"logLevel"`
	License            bool   `json:"license"`
}

func validateArguments(args *Arguments) error {
	if args.WizClientID == "" {
		return errors.New("WizClientID is required")
	}
	if args.WizClientSecret == "" {
		return errors.New("WizClientSecret is required")
	}
	if args.WizQueryURL == "" {
		return errors.New("WizQueryURL is required")
	}
	if args.WizAuthURL == "" {
		return errors.New("WizAuthURL is required")
	}
	if args.ScanSubscriptionID == "" {
		return errors.New("ScanSubscriptionID is required")
	}
	if args.ScanCloudType == "" {
		return errors.New("ScanCloudType is required")
	}
	if args.ScanProviderID == "" {
		return errors.New("ScanProviderID is required")
	}

	return nil
}

func saveConfig(config *Arguments, filePath string) error {
	config.Install = false // Ensure Install is always false when saving config

	// Get the directory path from the file path
	dir := filepath.Dir(filePath)

	// Check if the directory exists, if not, create it
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Marshal the config struct to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Encode the JSON data using base64
	encodedData := base64.StdEncoding.EncodeToString(data)

	// Convert encoded data to byte slice for writing to file
	byteData := []byte(encodedData)

	// Write the base64 encoded data to the file with 0600 permissions to ensure the file is only accessible to the user
	if err := os.WriteFile(filePath, byteData, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func readConfig(filePath string, config *Arguments) error {
	encodedData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	// Decode the base64-encoded data
	decodedData, err := base64.StdEncoding.DecodeString(string(encodedData))
	if err != nil {
		return fmt.Errorf("failed to decode base64 data: %w", err)
	}
	// Unmarshal the JSON data into the Arguments struct
	if err = json.Unmarshal(decodedData, config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

func ProcessArguments() (*Arguments, error) {
	args := &Arguments{}
	var configFilePath string
	var logLevel string

	flag.StringVar(&logLevel, "logLevel", "info", "Set log level (info, error, etc.)")
	flag.StringVar(&args.WizClientID, "wizClientId", "", "Wiz Client ID")
	flag.StringVar(&args.WizClientSecret, "wizClientSecret", "", "Wiz Client Secret")
	flag.StringVar(&args.WizQueryURL, "wizQueryUrl", "", "Wiz Query URL")
	flag.StringVar(&args.WizAuthURL, "wizAuthUrl", "", "Wiz Auth URL")
	flag.StringVar(&args.ScanSubscriptionID, "scanSubscriptionId", "", "Scan Subscription ID")
	flag.StringVar(&args.ScanCloudType, "scanCloudType", "", "Scan Cloud Type")
	flag.StringVar(&args.ScanProviderID, "scanProviderId", "", "Scan Provider ID")
	flag.BoolVar(&args.Save, "save", false, "Set to true to save the configuration (ignored if install flag is set)")
	flag.StringVar(&configFilePath, "config", "config.json", "Path to the configuration file (ignored if install flag is set)")
	flag.BoolVar(&args.Install, "install", false, "Install the application")
	flag.BoolVar(&args.Uninstall, "uninstall", false, "Uninstall the application")
	flag.BoolVar(&args.License, "license", false, "Print License and Support Information")

	flag.Parse()

	args.LogLevel = logLevel

	// Print Support info
	if args.License {
		fmt.Println(`
		By using this software and associated documentation files (the “Software”) you hereby agree and understand that:
		- The use of the Software is free of charge and may only be used by Wiz customers for its internal purposes.
		- The Software should not be distributed to third parties.
		- The Software is not part of Wiz’s Services and is not subject to your company’s services agreement with Wiz.
	  
		THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
		TO WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL WIZ
		BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
		ARISING FROM, OUT OF OR IN CONNECTION WITH THE USE OF THIS SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
		`)
		os.Exit(0)
	}

	// Enforce mutual exclusivity
	if args.Install && args.Uninstall {
		return nil, errors.New("'-install' and '-uninstall' cannot be used together")
	}

	// If uninstall is requested, we can immediately return since no other flags are needed
	if args.Uninstall {
		return args, nil
	}

	currInstall := args.Install
	// If the save option isn't flagged
	if !args.Save {

		// Validate the arguments
		if err := validateArguments(args); err != nil {
			// If validation fails, attempt to load arguments from the config file
			if err := readConfig(configFilePath, args); err != nil {
				// If loading fails, return an error
				return nil, fmt.Errorf("error reading config file: %v", err)
			}
			// If loading succeeds, validate arguments again
			if err := validateArguments(args); err != nil {
				// If re-validation fails, return an error
				return nil, fmt.Errorf("error validating arguments: %v", err)
			}
		}
	}

	if args.Save {
		if err := validateArguments(args); err != nil {
			return nil, fmt.Errorf("error validating arguments: %v", err)
		}
		if err := saveConfig(args, configFilePath); err != nil {
			return nil, fmt.Errorf("error saving config: %v", err)
		}
	}

	args.Install = currInstall
	return args, nil
}
