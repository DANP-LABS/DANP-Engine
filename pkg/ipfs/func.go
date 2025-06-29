package ipfs

import "os"

// GetDATAFromIPFSCID retrieves the data from an IPFS CID, extracts the WASM files, and returns them as byte slices.
//
// Parameters:
//   - ipfsC: An instance of the IPFSClient struct, which provides methods for interacting with the IPFS network.
//   - cid: The CID (Content Identifier) of the data to be retrieved from the IPFS network.
//
// Returns:
//   - D: A slice of byte slices, each representing the content of a WASM file extracted from the CID.
//   - err: An error, if any, encountered during the retrieval or extraction process.
//
// Note: This function assumes that the CID contains WASM files and that the extracted files are stored in a temporary directory.
func GetDATAFromIPFSCID(ipfsC *IPFSClient, cid string) (D [][]byte, err error) {
	// Retrieve the data from the given CID using the provided IPFSClient instance.
	data, err := ipfsC.GetDataFromCID(cid)
	if err != nil {
		return nil, err
	}

	// Create a temporary file to store the CAR (IPFS Content Addressing: References) data.
	fcar, err := os.CreateTemp("", cid+"Car")
	if err != nil {
		return nil, err
	}

	// Defer the removal of the temporary file when the function exits.
	defer os.Remove(fcar.Name())

	// Write the retrieved data to the temporary file.
	_, err = fcar.Write(data)
	if err != nil {
		return nil, err
	}
	fcar.Close()

	// Create a temporary directory to store the extracted files.
	wasmdname, err := os.MkdirTemp("", cid+"CarExtractOutputDir")
	if err != nil {
		return nil, err
	}

	// Defer the removal of the temporary directory when the function exits.
	defer os.RemoveAll(wasmdname)

	// Extract the files from the CAR data to the temporary directory.
	_, err = ExtractCarFile(fcar.Name(), wasmdname)
	if err != nil {
		return nil, err
	}

	// Read the contents of the extracted files in the temporary directory.
	entries, err := os.ReadDir(wasmdname)
	if err != nil {
		return nil, err
	}

	// Iterate over the contents of the temporary directory.
	for _, e := range entries {
		// Skip directories.
		if e.IsDir() {
			continue
		}

		// Get the path of the current file.
		wasmPath := wasmdname + "/" + e.Name()

		// Read the content of the current file as bytes.
		wasmdata, err := os.ReadFile(wasmPath)
		if err != nil {
			return nil, err
		}

		// Append the content of the current file to the slice of byte slices.
		D = append(D, wasmdata)
	}

	return
}
