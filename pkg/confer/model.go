package confer

import (
	"sync"
)

// Confer represents the configuration structure with a read-write mutex for synchronization.
type Confer struct {
	Mutex sync.RWMutex // Mutex for read-write synchronization
	Opts  confS        // Configuration options
}

// confS holds various configuration options parsed from a YAML file.
type confS struct {
	ApptypeConf      string           `yaml:"app-type"`           // Application type: WASM_WORKER
	NetModelConf     string           `yaml:"net-model"`          // Network model for HTTP handling: NetPoll(gin) or RAWEPOLL(fiber)
	NetWork          NetWorkS         `yaml:"NetWork"`            // Configuration for handling incoming network traffic
	DebugConf        DebugConfS       `yaml:"debug"`              // Configuration for runtime memory debugging
	WASMModulesFiles WASMModulesFileS `yaml:"wasm-modules-files"` // Configuration for WASM modules loaded from file paths
	WASMModulesIPFS  WASMModulesIPFSS `yaml:"wasm-modules-ipfs"`  // Configuration for WASM modules loaded from IPFS
}

// WASMModulesFileS holds configuration for WASM modules loaded from file paths.
type WASMModulesFileS struct {
	Enable        bool     `yaml:"enable"` // Whether to enable loading WASM modules from files
	WASMFilePaths []string `yaml:"path"`   // List of file paths for WASM modules
}

// LassieNetS holds network configuration for Lassie.
type LassieNetS struct {
	Scheme string `yaml:"scheme"` // Network scheme (e.g., http, https)
	Host   string `yaml:"host"`   // Host address
	Port   int    `yaml:"port"`   // Port number
}

// WASMModulesIPFSS holds configuration for WASM modules loaded from IPFS.
type WASMModulesIPFSS struct {
	Enable    bool       `yaml:"enable"`     // Whether to enable loading WASM modules from IPFS
	LassieNet LassieNetS `yaml:"lassie-net"` // Network configuration for Lassie
	CIDS      []string   `yaml:"cids"`       // List of CIDs for WASM modules
}

// NetWorkS is used to handle incoming traffic network configuration.
type NetWorkS struct {
	BindNetWork  string `yaml:"bind-network"`  // Network transport layer type: TCP or UDP
	ProtocolType string `yaml:"protocol-type"` // Application layer network protocol: HTTP, RESP, or QUIC
	BindAddress  string `yaml:"bind-address"`  // Network listening address, where the application will listen for incoming traffic
}

// DebugConfS holds debug configuration options.
type DebugConfS struct {
	Enable        bool   `yaml:"enable"`          // Whether to enable debugging
	PprofBindAddr string `yaml:"pprof-bind-addr"` // Address for performance analysis network binding
}
