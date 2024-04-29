package main

import (
	"runtime"

	"github.com/jtb75/wiz-scan/pkg/utilities"
)

func main() {

	// Get the detected operating system
	operatingSystem := runtime.GOOS

	// Print the detected operating system
	utilities.Log.Info("Operating System:", operatingSystem)

}
