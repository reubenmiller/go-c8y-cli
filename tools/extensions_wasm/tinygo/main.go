package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

//go:wasmimport env current_year
func getYear() uint32

//go:wasmimport env check
func check(uint32)

type Options struct {
	Name    string `json:"name"`
	Default string `json:"default,omitempty"`
}

//go:wasmexport help
func help() {
	fmt.Printf(`
c8y-tinygo [OPTIONS] [command]
`)
}

//go:wasmexport options
func options() {
	config := Options{
		Name:    "hello",
		Default: "again",
	}

	b, err := json.Marshal(config)
	if err != nil {
		os.Exit(1)
	}
	fmt.Printf("%s", b)
}

func sleepy(v uint64) {
	i := uint64(0)
	j := v / 100

	for i < j {
		time.Sleep(100 * time.Millisecond)
		i = i + 1
	}
}

//go:wasmexport run
func run() {
	tenant := os.Getenv("C8Y_TENANT")
	check(1234)

	sleepy(5000)

	fmt.Printf("Current Year: %d\n", getYear())

	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		println("could not read input", err)
		os.Exit(1)
	}

	fmt.Printf("Stdin:\n%s\n\n", string(input))

	println("tenant=", tenant)
	i := 0
	for i < 10 {
		fmt.Printf("Hello word: i=%d\n", i)
		i += 1
	}
}

// main is required for the `wasip1` target, even if it isn't used.
func main() {}
