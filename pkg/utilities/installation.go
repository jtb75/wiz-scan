package utilities

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// CopyFile copies a file from src to dst.
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}

// InstallAndScheduleTask installs the application and sets up a scheduled task to run it daily.
func InstallAndScheduleTask(args *Arguments) error {
	// Determine the appropriate "Program Files" directory.
	programFilesDir := os.Getenv("ProgramFiles")
	if programFilesDir == "" {
		return fmt.Errorf("failed to determine Program Files directory")
	}

	// Create the "Wiz-Scan" directory within the "Program Files" directory.
	wizScanDir := filepath.Join(programFilesDir, "Wiz-Scan")
	err := os.MkdirAll(wizScanDir, 0755)
	if err != nil {
		return err
	}
	// Copy the application executable into the "Wiz-Scan" directory.
	executablePath, err := os.Executable()
	if err != nil {
		return err
	}
	executableName := filepath.Base(executablePath)
	destinationPath := filepath.Join(wizScanDir, executableName)
	err = CopyFile(executablePath, destinationPath)
	if err != nil {
		return err
	}

	// Save configuration to file
	configFilePath := filepath.Join(wizScanDir, "config.txt")
	err = saveConfig(args, configFilePath)
	if err != nil {
		return fmt.Errorf("error saving configuration: %v", err)
	}

	// Concatenate the destinationPath with the -config option and the configFilePath
	taskCommand := fmt.Sprintf("%s -config %s", destinationPath, configFilePath)

	// Use schtasks command to create a scheduled task.
	err = createScheduledTask("WizScanTask", taskCommand)
	if err != nil {
		return fmt.Errorf("error creating scheduled task: %v", err)
	}

	return nil
}

// RemoveAll removes a directory and all its contents recursively.
func RemoveAll(dir string) error {
	err := os.RemoveAll(dir)
	if err != nil {
		return err
	}
	return nil
}

// RemoveScheduledTask deletes the scheduled task created by InstallAndScheduleTask.
func RemoveScheduledTask() error {
	// Use schtasks command to delete the scheduled task.
	cmd := exec.Command("schtasks", "/Delete", "/TN", "WizScanTask", "/F")
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// UninstallAndRemoveTask undoes the changes made by InstallAndScheduleTask.
func UninstallAndRemoveTask() error {
	// Determine the appropriate "Program Files" directory.
	programFilesDir := os.Getenv("ProgramFiles")
	if programFilesDir == "" {
		return fmt.Errorf("failed to determine Program Files directory")
	}

	// Delete the "Wiz-Scan" directory and its contents.
	wizScanDir := filepath.Join(programFilesDir, "Wiz-Scan")
	err := RemoveAll(wizScanDir)
	if err != nil {
		return err
	}

	// Remove the scheduled task.
	err = RemoveScheduledTask()
	if err != nil {
		return err
	}

	return nil
}

func createScheduledTask(taskName, destinationPath string) error {
	taskExists, err := taskExists(taskName)
	if err != nil {
		return err
	}
	if taskExists {
		fmt.Printf("Task '%s' already exists\n", taskName)
		return nil
	}

	// Generate random start time between 8:00 PM and 4:00 AM
	startTime := time.Date(0, 1, 1, 20, 0, 0, 0, time.UTC) // Start time: 8:00 PM
	randomMinutes, err := randInt(480)                     // Random number of minutes between 0 and 480 (8 hours)
	if err != nil {
		return fmt.Errorf("error generating random number: %v", err)
	}
	startTime = startTime.Add(time.Duration(randomMinutes) * time.Minute)

	// Format start time as HH:mm
	startTimeFormatted := startTime.Format("15:04")

	cmd := exec.Command("schtasks", "/Create", "/SC", "DAILY", "/TN", taskName, "/TR", destinationPath, "/ST", startTimeFormatted)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error creating scheduled task: %v", err)
	}
	fmt.Printf("Scheduled task '%s' created successfully with start time: %s\n", taskName, startTimeFormatted)
	return nil
}

func taskExists(taskName string) (bool, error) {
	cmd := exec.Command("schtasks", "/Query", "/TN", taskName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "ERROR: The system cannot find the file specified.") {
			// Task doesn't exist
			return false, nil
		}
		return false, fmt.Errorf("error checking task existence: %v", err)
	}
	// Task exists
	return true, nil
}

// randInt generates a random integer between 0 and max.
func randInt(max int) (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()), nil
}
