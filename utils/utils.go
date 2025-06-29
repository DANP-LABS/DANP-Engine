package utils

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"time"
)

// IsFileExist checks if a file exists at the specified path.
// It returns true if the file exists, false otherwise.
func IsFileExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		// Log the error if it's not a "file does not exist" error
		log.Printf("Error checking file existence: %v", err)
	}
	return true
}

// ReadFileData reads the contents of a file specified by the FileURI parameter.
// It returns the file data as a byte slice and an error if the file does not exist or if there was an error reading the file.
func ReadFileData(FileURI string) ([]byte, error) {
	if !IsFileExist(FileURI) {
		return nil, fmt.Errorf("%s does not exist", FileURI)
	}

	fileData, err := os.ReadFile(FileURI)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return fileData, nil
}

// InArray checks if a string is present in an array of strings.
func InArray(in string, array []string) bool {
	for _, element := range array {
		if in == element {
			return true
		}
	}
	return false
}

// GoWithRecover wraps a `go func()` with recover() to handle panics gracefully.
// It logs the panic details and optionally calls a recover handler function.
func GoWithRecover(handler func(), recoverHandler func(r interface{})) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("%s goroutine panic: %v\n%s\n", time.Now().Format("2006-01-02 15:04:05"), r, string(debug.Stack()))
				if recoverHandler != nil {
					go func() {
						defer func() {
							if p := recover(); p != nil {
								log.Printf("recover goroutine panic: %v\n%s\n", p, string(debug.Stack()))
							}
						}()
						recoverHandler(r)
					}()
				}
			}
		}()
		handler()
	}()
}
