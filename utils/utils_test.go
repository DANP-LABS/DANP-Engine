package utils

import (
	"os"
	"testing"
)

func TestInArray(t *testing.T) {
	// Test case 1: In value is present in the array
	array := []string{"apple", "banana", "cherry"}
	in := "banana"
	expected := true
	result := InArray(in, array)
	if result != expected {
		t.Errorf("Expected %v, but got %v", expected, result)
	}

	// Test case 2: In value is not present in the array
	array = []string{"apple", "banana", "cherry"}
	in = "mango"
	expected = false
	result = InArray(in, array)
	if result != expected {
		t.Errorf("Expected %v, but got %v", expected, result)
	}

	// Test case 3: Empty array
	array = []string{}
	in = "apple"
	expected = false
	result = InArray(in, array)
	if result != expected {
		t.Errorf("Expected %v, but got %v", expected, result)
	}

	// Test case 4: Empty in value
	array = []string{"apple", "banana", "cherry"}
	in = ""
	expected = false
	result = InArray(in, array)
	if result != expected {
		t.Errorf("Expected %v, but got %v", expected, result)
	}

	// Test case 5: Array with duplicate values
	array = []string{"apple", "banana", "cherry", "banana"}
	in = "banana"
	expected = true
	result = InArray(in, array)
	if result != expected {
		t.Errorf("Expected %v, but got %v", expected, result)
	}
}

func TestIsFileExist(t *testing.T) {
	// Setup
	fileName := "testfile.txt"
	nonexistentFileName := "nonexistent.txt"

	// Create a test file
	_, err := os.Create(fileName)
	if err != nil {
		t.Fatalf("Failed to set up test file: %v", err)
	}

	// Test cases
	tests := []struct {
		name     string
		filePath string
		want     bool
	}{
		{
			name:     "File exists",
			filePath: fileName,
			want:     true,
		},
		{
			name:     "File does not exist",
			filePath: nonexistentFileName,
			want:     false,
		},
		{
			name:     "Empty file path",
			filePath: "",
			want:     false,
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsFileExist(tt.filePath); got != tt.want {
				t.Errorf("IsFileExist() = %v, want %v", got, tt.want)
			}
		})
	}

	// Teardown
	err = os.Remove(fileName)
	if err != nil {
		t.Fatalf("Failed to tear down test file: %v", err)
	}
}

func TestReadFileData(t *testing.T) {
	// Setup
	fileContent := "This is some test content"
	tempFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write([]byte(fileContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Test case: valid file
	data, err := ReadFileData(tempFile.Name())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if string(data) != fileContent {
		t.Errorf("Expected %q, got %q", fileContent, string(data))
	}

	// Test case: non-existent file
	_, err = ReadFileData("non-existent-file")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestReadFileDataWithEmptyString(t *testing.T) {
	// Test case: empty string as file URI
	_, err := ReadFileData("")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}
