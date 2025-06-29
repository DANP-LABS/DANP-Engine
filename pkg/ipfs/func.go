package ipfs

import (
	"fmt"
	"os"
)

// ExtractWASMFromCID retrieves WASM files from IPFS by CID and returns them as byte slices.
// client: Configured IPFS client
// cid: Content Identifier of the data to retrieve
// Returns: Slice of WASM file contents and any error encountered
func ExtractWASMFromCID(client *Client, cid string) ([][]byte, error) {
	// Retrieve CAR data from IPFS
	data, err := client.Retrieve(cid)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve CID: %w", err)
	}

	// Create temp file for CAR data
	carFile, err := os.CreateTemp("", cid+"Car")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(carFile.Name())

	// Write CAR data to temp file
	if _, err := carFile.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write CAR data: %w", err)
	}
	carFile.Close()

	// Create temp directory for extraction
	extractDir, err := os.MkdirTemp("", cid+"CarExtractOutputDir")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(extractDir)

	// Extract CAR file contents
	if _, err := ExtractCarFile(carFile.Name(), extractDir); err != nil {
		return nil, fmt.Errorf("failed to extract CAR file: %w", err)
	}

	// Read all extracted files
	entries, err := os.ReadDir(extractDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read extract dir: %w", err)
	}

	var wasmFiles [][]byte
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := extractDir + "/" + entry.Name()
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", entry.Name(), err)
		}

		wasmFiles = append(wasmFiles, content)
	}

	return wasmFiles, nil
}
