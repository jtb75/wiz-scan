package utilities

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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
func InstallAndScheduleTask() error {
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

	// Use schtasks command to create a scheduled task.
	cmd := exec.Command("schtasks", "/Create", "/SC", "DAILY", "/TN", "WizScanTask", "/TR", destinationPath, "/ST", "00:00")
	err = cmd.Run()
	if err != nil {
		return err
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
