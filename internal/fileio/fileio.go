package fileio

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func WriteToRelativePath(filePath string, data []byte) {
	// Write the byte slice to a file
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Combine the current directory with the file path
	filePath = filepath.Clean(filePath)
	fullPath := filepath.Join(dir, filePath)

	// Extract the directory from the full path
	dirPath := filepath.Dir(fullPath)

	// Create the directory if it doesn't exist
	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		log.Fatal(err)
	}

	// Write the file
	err = os.WriteFile(fullPath, data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func WriteToAbsolutePath(filePath string, data []byte) {
	// Write the byte slice to a file

	// Combine the current directory with the file path
	filePath = filepath.Clean(filePath)

	// Extract the directory from the full path
	dirPath := filepath.Dir(filePath)

	// Create the directory if it doesn't exist
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		log.Fatal(err)
	}

	// Write the file
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
