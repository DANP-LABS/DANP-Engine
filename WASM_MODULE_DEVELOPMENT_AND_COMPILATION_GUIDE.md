# DANP-Engine WASM Module Development Manual

This document provides a clear guide for developers on creating, compiling, and integrating new WASM modules into the DANP-Engine.

## Introduction

WASM (WebAssembly) modules are a core component of the DANP-Engine, offering a secure and portable way to execute decentralized AI tools and various computational tasks. By encapsulating functionality within WASM modules, we ensure that code runs in a sandboxed environment, protecting the host system's security.

## Technology Stack

To streamline the development process and maintain consistency across the project, we have chosen the following technologies for writing WASM modules:

-   **Go:** Aligns with the core language of the DANP-Engine.
-   **extism/go-pdk:** A powerful Go PDK (Plugin Development Kit) that greatly simplifies WASM module development, especially input/output handling and interaction with the host environment, avoiding complex manual memory management.
-   **TinyGo:** A Go compiler for embedded systems and WebAssembly, which produces smaller WASM files.

## Steps to Write a New Module

Here are the complete steps to add a new WASM module to the DANP-Engine:

### 1. Create the Directory

In the `wasm-examples/` directory at the project root, create a new subdirectory for your module. For example, for the `say_hello` module:

```bash
mkdir -p wasm-examples/say_hello
```

### 2. Write the Go Code

In your module's directory, create a `main.go` file. Use `extism/go-pdk` to handle inputs and outputs.

**Example: `wasm-examples/say_hello/main.go`**

```go
package main

import (
	"fmt"
	"github.com/extism/go-pdk"
)

//export say_hello
func say_hello() int32 {
	// Read the input string from the host
	name := pdk.InputString()

	// Create the greeting message
	greeting := fmt.Sprintf("Hello, %s!", name)

	// Return the greeting to the host
	pdk.OutputString(greeting)
	return 0 // Indicate success
}

func main() {}
```

### 3. Compile the WASM Module

We provide a unified command for compiling, which handles Go module initialization, dependency fetching, and WASM compilation automatically. To produce the smallest possible WASM files, we use several optimization flags.

Run the following command in your module's directory:

```bash
cd wasm-examples/YOUR_MODULE_NAME && \
go mod init YOUR_MODULE_NAME >/dev/null 2>&1 && \
go get github.com/extism/go-pdk && \
GOOS=wasip1 GOARCH=wasm tinygo build -o YOUR_MODULE_NAME.wasm -target wasi -opt=z -no-debug -scheduler=none main.go
```

Replace `YOUR_MODULE_NAME` with the name of your module.

**Optimization Flags:**
-   `-opt=z`: Aggressively optimizes for size, potentially at the cost of some execution speed.
-   `-no-debug`: Strips all debug information from the binary.
-   `-scheduler=none`: Removes the Go scheduler, which is not needed for simple, non-concurrent WASM modules.

### 4. Integrate into `mcp_manifest.yaml`


Finally, in the `config/mcp_manifest.yaml` file, add your new module to the `modules` list.

**Example:**

```yaml
modules:
  - name: "say_hello"
    wasm_path: "file://wasm-examples/say_hello/say_hello.wasm"
    tools:
      - name: "say_hello"
        description: "Greet someone by name"
        inputs:
          - name: "name"
            type: "string"
            required: true
            description: "Name to greet"
        outputs:
          type: "string"
          description: "Greeting message"
```

## Notes

-   **Input/Output:** `extism/go-pdk` greatly simplifies I/O. For simple strings, you can use `pdk.InputString()` and `pdk.OutputString()`. For complex JSON data, read the bytes with `pdk.Input()`, then parse with `json.Unmarshal`.
-   **Error Handling:** It is recommended to add proper error handling in your WASM functions and pass error messages back to the host via return values or output strings.
-   **Performance:** Keep your WASM modules lightweight and efficient. Avoid long-running or memory-intensive tasks within the WASM module itself.
