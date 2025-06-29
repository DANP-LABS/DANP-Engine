package confer

import (
	"reflect"
	"testing"
)

// Unit tests for parseYamlFromBytes function
func TestParseYamlFromBytes(t *testing.T) {
	// Test case 1: Valid YAML data
	validYAML := []byte(`
app-type: WASM_WORKER
net-model: NetPoll
NetWork:
  bind-network: TCP
  protocol-type: HTTP
  bind-address: 0.0.0.0:8080
debug:
  enable: true
  pprof-bind-addr: 0.0.0.0:6060
wasm-modules-files:
  enable: true
  path:
    - /path/to/module1.wasm
    - /path/to/module2.wasm
wasm-modules-ipfs:
  enable: true
  lassie-net:
    scheme: http
    host: localhost
    port: 5001
  cids:
    - QmExampleCID1
    - QmExampleCID2
`)
	expectedConf := confS{
		ApptypeConf:  "WASM_WORKER",
		NetModelConf: "NetPoll",
		NetWork: NetWorkS{
			BindNetWork:  "TCP",
			ProtocolType: "HTTP",
			BindAddress:  "0.0.0.0:8080",
		},
		DebugConf: DebugConfS{
			Enable:        true,
			PprofBindAddr: "0.0.0.0:6060",
		},
		WASMModulesFiles: WASMModulesFileS{
			Enable:        true,
			WASMFilePaths: []string{"/path/to/module1.wasm", "/path/to/module2.wasm"},
		},
		WASMModulesIPFS: WASMModulesIPFSS{
			Enable: true,
			LassieNet: LassieNetS{
				Scheme: "http",
				Host:   "localhost",
				Port:   5001,
			},
			CIDS: []string{"QmExampleCID1", "QmExampleCID2"},
		},
	}

	parsedConf, err := parseYamlFromBytes(validYAML)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !reflect.DeepEqual(parsedConf, expectedConf) {
		t.Errorf("Parsed configuration does not match expected configuration")
	}

	// Test case 2: Empty YAML data
	emptyYAML := []byte("")
	_, err = parseYamlFromBytes(emptyYAML)
	if err == nil {
		t.Errorf("Expected error for empty YAML data, but got none")
	}

	// Test case 3: Invalid YAML data
	invalidYAML := []byte(`
invalid: yaml: data
`)
	_, err = parseYamlFromBytes(invalidYAML)
	if err == nil {
		t.Errorf("Expected error for invalid YAML data, but got none")
	}
}
