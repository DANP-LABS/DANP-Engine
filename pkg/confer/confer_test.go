package confer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitConfigWithSampleConfig(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "confer_test")
	assert.NoError(t, err, "Failed to create temporary directory")
	defer os.RemoveAll(tempDir) // Clean up after test

	// Create a sample configuration file in the temporary directory
	sampleConfFilePath := filepath.Join(tempDir, "sample_config.yaml")
	sampleConfigContent := `
app-type: "WASM-IPFS-Serverless-WORKER"
net-model: "NETPOLL"
NetWork:
  bind-network: "TCP"
  protocol-type: "HTTP"
  bind-address: "127.0.0.1:28080"
debug:
  enable: false
  pprof-bind-addr: "127.0.0.1:19090"
wasm-modules-files:
  enable: false
  path:
    - "hello.wasm"
wasm-modules-ipfs:
  enable: false
  lassie-net:
    scheme: "http"
    host: "x.x.x.x"
    port: 5001
  cids:
    - "QmeDsaLTc8dAfPrQ5duC4j5KqPdGbcinEo5htDqSgU8u8Z"
`
	err = os.WriteFile(sampleConfFilePath, []byte(sampleConfigContent), 0644)
	assert.NoError(t, err, "Failed to create sample configuration file")

	// Test case 1: Valid configuration file
	conf, err := InitConfig("testApp", sampleConfFilePath)
	assert.NoError(t, err, "Expected no error for valid configuration file")
	assert.NotNil(t, conf, "Expected a non-nil Confer instance")

	// Verify the parsed configuration values
	assert.Equal(t, "WASM-IPFS-Serverless-WORKER", conf.Opts.ApptypeConf, "Expected ApptypeConf to be set correctly")
	assert.Equal(t, "NETPOLL", conf.Opts.NetModelConf, "Expected NetModelConf to be set correctly")
	assert.Equal(t, "TCP", conf.Opts.NetWork.BindNetWork, "Expected BindNetwork to be set correctly")
	assert.Equal(t, "HTTP", conf.Opts.NetWork.ProtocolType, "Expected ProtocolType to be set correctly")
	assert.Equal(t, "127.0.0.1:28080", conf.Opts.NetWork.BindAddress, "Expected BindAddress to be set correctly")
	assert.False(t, conf.Opts.DebugConf.Enable, "Expected DebugConf.Enable to be false")
	assert.Equal(t, "127.0.0.1:19090", conf.Opts.DebugConf.PprofBindAddr, "Expected PprofBindAddr to be set correctly")
	assert.Equal(t, "http", conf.Opts.WASMModulesIPFS.LassieNet.Scheme, "Expected Scheme to be set correctly")
	assert.Equal(t, "x.x.x.x", conf.Opts.WASMModulesIPFS.LassieNet.Host, "Expected Host to be set correctly")
	assert.Equal(t, 5001, conf.Opts.WASMModulesIPFS.LassieNet.Port, "Expected Port to be set correctly")
	assert.Equal(t, []string{"QmeDsaLTc8dAfPrQ5duC4j5KqPdGbcinEo5htDqSgU8u8Z"}, conf.Opts.WASMModulesIPFS.CIDS, "Expected Cids to be set correctly")

	// Test case 2: Environment variable override
	os.Setenv("NetModelConf", "RAWEPOLL")
	defer os.Unsetenv("NetModelConf")

	conf, err = InitConfig("testApp", sampleConfFilePath)
	assert.NoError(t, err, "Expected no error for valid configuration file with env override")
	assert.Equal(t, "NETPOLL", conf.Opts.NetModelConf, "Expected NetModelConf to be overridden by environment variable")
}
