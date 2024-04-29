package main

import (
	"os"
	"os/exec"
	"runtime"

	"github.com/jtb75/wiz-scan/pkg/utilities"
)

func main() {

	// Get the detected operating system
	operatingSystem := runtime.GOOS

	// Print the detected operating system
	utilities.Log.Info("Operating System:", operatingSystem)

	// Perform OS-specific operations
	if operatingSystem == "windows" {
		// Windows-specific code
		utilities.Log.Info("Running on Windows")
		cmd := exec.Command("cmd", "/c", "dir") // Example Windows command
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	} else if operatingSystem == "linux" {
		// Linux-specific code
		utilities.Log.Info("Running on Linux")
		cmd := exec.Command("ls", "-l") // Example Linux command
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	} else {
		// Unsupported operating system
		utilities.Log.Warn("Unsupported operating system:", operatingSystem)
	}
}
