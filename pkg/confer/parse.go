package confer

import (
	"errors"
	"fmt"

	utils "github.com/IceFireLabs/DANP-Engine/utils"
	"gopkg.in/yaml.v2"
)

// parseYamlFromBytes parses YAML data from a byte slice.
// It returns the parsed configuration data and any error encountered during parsing.
func parseYamlFromBytes(originData []byte) (data confS, err error) {
	// Check if the input data is empty
	if len(originData) == 0 {
		err = errors.New("yaml source data is empty")
		return
	}

	// Unmarshal the YAML data into the confS struct
	err = yaml.Unmarshal(originData, &data)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal YAML: %v", err)
		return
	}

	return
}

// parseYamlFromFile parses YAML data from a file specified by its path.
// It returns the parsed configuration data and any error encountered during the process.
func parseYamlFromFile(filePath string) (confS, error) {
	// Read the file data from the specified path
	fileData, err := utils.ReadFileData(filePath)
	if err != nil {
		return confS{}, err
	}

	// Parse the YAML data from the file data
	return parseYamlFromBytes(fileData)
}
