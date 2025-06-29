package ipfs

import (
	"errors"
	"os"

	"github.com/ipfs/go-cid"
	carstorage "github.com/ipld/go-car/v2/storage"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/storage"
)

// ExtractCarFile extracts files from a Car (Content Addressable References) file.
// It reads the Car file at the specified path, extracts the files rooted at the provided CIDs,
// and saves them to the specified output directory.
//
// The function returns the total number of extracted files and an error if any.
// If no files are extracted, it returns an error indicating so.
//
// Parameters:
//   - carfilePath: The path to the Car file to be extracted.
//   - outputDir: The directory where the extracted files will be saved.
//
// Returns:
//   - extractedFiles: The total number of extracted files.
//   - err: An error if any occurred during the extraction process.
func ExtractCarFile(carfilePath string, outputDir string) (extractedFiles int, err error) {
	// Initialize the storage variable to nil.
	var store storage.ReadableStorage

	// Initialize an empty slice for the roots of the Car file.
	var roots []cid.Cid

	// Open the Car file at the specified path.
	carFile, err := os.Open(carfilePath)
	if err != nil {
		// Return an error if the file cannot be opened.
		return 0, err
	}

	// Defer the closing of the Car file.
	defer carFile.Close()

	// Open the Car file for reading and initialize the storage variable with the readable Car storage.
	store, err = carstorage.OpenReadable(carFile)
	if err != nil {
		// Return an error if the Car file cannot be read.
		return 0, err
	}

	// Get the roots of the Car file.
	roots = store.(carstorage.ReadableCar).Roots()

	// Initialize the link system with default settings and set the read storage to the Car file's storage.
	ls := cidlink.DefaultLinkSystem()
	ls.TrustedStorage = true
	ls.SetReadStorage(store)

	// Iterate over the roots of the Car file.
	for _, root := range roots {
		// Extract the files rooted at the current CID and increment the count of extracted files.
		count, err := extractRoot(&ls, root, outputDir, nil)
		if err != nil {
			// Return an error if any occurs during the extraction process.
			return 0, err
		}
		extractedFiles += count
	}

	// If no files are extracted, return an error.
	if extractedFiles == 0 {
		return 0, errors.New("no files extracted")
	}

	// Return the total number of extracted files and nil if no errors occurred.
	return extractedFiles, nil
}
