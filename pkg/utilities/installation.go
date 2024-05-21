package utilities

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

	// Get file info for the source file to determine its permissions
	fileInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Get the mode (permissions) from the source file
	mode := fileInfo.Mode()

	// Set the same mode (permissions) for the destination file
	err = os.Chmod(dst, mode)
	if err != nil {
		return err
	}

	return nil
}

// InstallAndScheduleTask installs the application and sets up a scheduled task to run it daily.
func InstallAndScheduleTask(args *Arguments) error {
	var programFilesDir, configFilePath, wizScanDir string

	// Determine the appropriate "Program Files" directory.
	if runtime.GOOS == "windows" {
		programFilesDir = os.Getenv("ProgramFiles")
		if programFilesDir == "" {
			return fmt.Errorf("failed to determine Program Files directory")
		}
		// Create the "Wiz-Scan" directory within the "Program Files" directory.
		wizScanDir = filepath.Join(programFilesDir, "Wiz-Scan")
		err := os.MkdirAll(wizScanDir, 0755)
		if err != nil {
			return err
		}
	} else {
		wizScanDir = "/usr/local/bin"
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
	if runtime.GOOS == "windows" {
		configFilePath = filepath.Join(wizScanDir, "config.txt")
	} else {
		configFilePath = "/etc/wiz-scan/config.txt"
	}
	err = saveConfig(args, configFilePath)
	if err != nil {
		return fmt.Errorf("error saving configuration: %v", err)
	}

	// Use schtasks command to create a scheduled task on Windows.
	if runtime.GOOS == "windows" {
		// Concatenate the destinationPath with the -config option and the configFilePath
		taskCommand := fmt.Sprintf("\"%s\" -config \"%s\"", destinationPath, configFilePath)
		err = createScheduledTask("WizScanTask", taskCommand)
		if err != nil {
			return fmt.Errorf("error creating scheduled task: %v", err)
		}
	} else {
		// Concatenate the destinationPath with the -config option and the configFilePath
		cronCommand := fmt.Sprintf("%s -config %s", destinationPath, configFilePath)
		// Install cron job on Linux
		err = installCronJob(cronCommand)
		if err != nil {
			return fmt.Errorf("error installing cron job: %v", err)
		}
	}

	return nil
}

func installCronJob(command string) error {
	// Check if the command already exists in the current crontab
	if commandExists("wiz-scan") {
		return fmt.Errorf("command '%s' already exists in the crontab", command)
	}

	// Generate random start time between 8:00 PM and 4:00 AM
	startTime := time.Date(0, 1, 1, 20, 0, 0, 0, time.UTC) // Start time: 8:00 PM

	// Create a new random number generator with a new source
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Generate random number of minutes between 0 and 480 (8 hours)
	randomMinutes := r.Intn(480)

	// Add the random number of minutes to the start time
	startTime = startTime.Add(time.Duration(randomMinutes) * time.Minute)

	// Convert startTime to cron format (minute hour format)
	minute := startTime.Minute()
	hour := startTime.Hour()

	// Setting up the cron job command with log redirection and configuration option
	cronJob := fmt.Sprintf("%d %d * * * %s >> /var/log/wiz-scan.log 2>&1\n", minute, hour, command)

	// Adding the cron job to the user's crontab
	cmd := exec.Command("bash", "-c", fmt.Sprintf("(crontab -l 2>/dev/null; echo '%s') | crontab -", cronJob))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to schedule cron job: %w", err)
	}
	return nil
}

func commandExists(command string) bool {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("crontab -l | grep -qF '%s'", command))
	err := cmd.Run()
	return err == nil
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
	var cmd *exec.Cmd

	// Use different commands based on the operating system
	if runtime.GOOS == "windows" {
		cmd = exec.Command("schtasks", "/Delete", "/TN", "WizScanTask", "/F")
	} else {
		// On Linux, remove the cron job
		return removeScheduledTaskLinux()
	}

	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// removeScheduledTaskLinux removes the scheduled cron job for the wizscan application.
func removeScheduledTaskLinux() error {
	// This command filters out 'wizscan' and any empty lines left behind.
	cmd := exec.Command("bash", "-c", "crontab -l | grep -v 'wiz-scan' | grep -v '^$' | crontab -")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove scheduled cron job: %w", err)
	}
	return nil
}

// UninstallAndRemoveTask undoes the changes made by InstallAndScheduleTask.
func UninstallAndRemoveTask() error {
	// Determine the appropriate directory paths based on the operating system
	var destinationPath, programFilesDir, configFilePath string
	if runtime.GOOS == "windows" {
		programFilesDir = os.Getenv("ProgramFiles")
		if programFilesDir == "" {
			return fmt.Errorf("failed to determine Program Files directory")
		}
		destinationPath = filepath.Join(programFilesDir, "Wiz-Scan")
	} else {
		executablePath, err := os.Executable()
		if err != nil {
			return err
		}
		executableName := filepath.Base(executablePath)
		wizScanDir := "/usr/local/bin"
		destinationPath = filepath.Join(wizScanDir, executableName)
		configFilePath = "/etc/wiz-scan/config.txt"
	}

	// Delete the "Wiz-Scan" directory and its contents.
	err := RemoveAll(destinationPath)
	if err != nil {
		return err
	}

	// Remove the config file
	if runtime.GOOS != "windows" {
		err = os.Remove(configFilePath)
		if err != nil {
			return err
		}
	}

	// Remove the scheduled task or cron job
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
		return nil
	}

	// Create a new random number generator
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Generate random start time between 8:00 PM and 4:00 AM
	startTime := time.Date(0, 1, 1, 20, 0, 0, 0, time.UTC) // Start time: 8:00 PM
	randomMinutes := r.Intn(480)                           // Random number of minutes between 0 and 480 (8 hours)
	startTime = startTime.Add(time.Duration(randomMinutes) * time.Minute)

	// Format start time as HH:mm
	startTimeFormatted := startTime.Format("15:04")

	// Use schtasks command to create a scheduled task.
	cmd := exec.Command("schtasks", "/Create", "/SC", "DAILY", "/TN", taskName, "/TR", destinationPath, "/ST", startTimeFormatted, "/F", "/NP")
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error creating scheduled task: %v", err)
	}
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
