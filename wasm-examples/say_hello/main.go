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
