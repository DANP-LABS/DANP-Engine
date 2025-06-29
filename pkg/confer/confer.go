package confer

import (
	"errors"
	"os"

	utils "github.com/IceFireLabs/DANP-Engine/utils"
	"github.com/go-chassis/go-archaius"
)

const (
	// Constants for network models
	NET_MODEL_NETPOLL  = "NETPOLL"
	NET_MODEL_RAWEPOLL = "RAWEPOLL"

	// Constant for application type
	APP_TYPE_WASM_WORKER = "WASM-IPFS-Serverless-WORKER"
)

// InitConfig initializes the configuration using archaius with the specified configuration file.
// It returns a pointer to the Confer instance and any error encountered during initialization.
func InitConfig(appName, confFilePath string) (*Confer, error) {
	// Initialize archaius with the required configuration file
	if err := archaius.Init(
		archaius.WithRequiredFiles([]string{confFilePath}),
	); err != nil {
		return nil, err
	}

	// Create a new Confer instance
	conf := &Confer{}

	// Unmarshal the configuration into conf.Opts
	if err := archaius.UnmarshalConfig(&conf.Opts); err != nil {
		return nil, err
	}

	// Return the pointer to the conf and nil error
	return conf, nil
}

var _confer *Confer

// GetNewConfer is the core function to get a Confer entity.
// It initializes the Confer instance with the specified app type and configuration file URI.
func GetNewConfer(appType, confFileURI string) (confer *Confer, err error) {
	// Return the existing Confer instance if it has already been initialized
	if _confer != nil {
		return _confer, nil
	}

	// Create a new Confer instance
	confer = &Confer{}

	// Parse configuration from the specified file
	confer.Opts, err = parseYamlFromFile(confFileURI)
	if err != nil {
		return nil, err
	}

	// Set the application type
	confer.Opts.ApptypeConf = appType

	// Configure network model options
	confer.replaceByEnv(&confer.Opts.NetModelConf)

	// Validate and set the default network model if necessary
	if !utils.InArray(confer.Opts.NetModelConf, []string{NET_MODEL_NETPOLL, NET_MODEL_RAWEPOLL}) {
		confer.Opts.NetModelConf = NET_MODEL_NETPOLL
	}

	// Check for Debug parameters
	if confer.Opts.DebugConf.Enable {
		// Ensure the pprof network listening address is not empty if debugging is enabled
		if len(confer.Opts.DebugConf.PprofBindAddr) == 0 {
			err = errors.New("pprof network listening address cannot be empty")
			return
		}
	}

	// Set the global Confer instance
	_confer = confer
	return _confer, nil
}

// replaceByEnv replaces the configuration value with the corresponding environment variable if it exists.
func (*Confer) replaceByEnv(confName *string) {
	// Get the value of the environment variable specified by 'confName'
	if s := os.Getenv(*confName); len(s) > 0 {
		// Update the value of 'confName' with the value from the environment variable
		*confName = s
	}
}

// Global returns the global Confer instance.
// If the global instance is nil, it returns a new instance of Confer.
func Global() *Confer {
	if _confer == nil {
		return &Confer{} // Return a new instance of Confer if _confer is nil
	}
	return _confer // Return the existing instance of Confer
}
