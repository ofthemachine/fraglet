package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ofthemachine/fraglet/pkg/runner"
)

const defaultFragletPath = "/FRAGLET"

func main() {
	flag.Usage = usage
	fragletPath := flag.String("fraglet-path", defaultFragletPath, "Path where code is mounted in container (default: /FRAGLET)")
	filePath := flag.String("file", "", "Path to code file (if not provided, reads from STDIN)")

	flag.Parse()

	if flag.NArg() < 1 {
		usage()
		os.Exit(1)
	}

	image := flag.Arg(0)
	var code string
	var err error

	// Read code from file if provided, otherwise from STDIN
	if *filePath != "" {
		codeBytes, err := os.ReadFile(*filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", *filePath, err)
			os.Exit(1)
		}
		code = string(codeBytes)
	} else {
		// Read from STDIN
		codeBytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from STDIN: %v\n", err)
			os.Exit(1)
		}
		code = string(codeBytes)
	}

	// Write code to temp file
	tmpFile, cleanup, err := writeTempFile(code)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temp file: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	// Create runner
	r := runner.NewRunner(image, "")

	// Execute with volume mount at fragletPath
	spec := runner.RunSpec{
		Container: image,
		Volumes: []runner.VolumeMount{
			{
				HostPath:      tmpFile,
				ContainerPath: *fragletPath,
				ReadOnly:      true,
			},
		},
	}

	ctx := context.Background()
	result, err := r.Run(ctx, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Execution failed: %v\n", err)
		os.Exit(1)
	}

	// Output results
	if result.Stdout != "" {
		fmt.Print(result.Stdout)
	}
	if result.Stderr != "" {
		fmt.Fprint(os.Stderr, result.Stderr)
	}

	// Exit with the same code as the execution
	os.Exit(result.ExitCode)
}

func writeTempFile(content string) (string, func(), error) {
	tmpFile, err := os.CreateTemp("", "fraglet-*")
	if err != nil {
		return "", nil, err
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", nil, err
	}

	tmpFile.Close()
	os.Chmod(tmpFile.Name(), 0644)

	absPath, _ := filepath.Abs(tmpFile.Name())
	cleanup := func() { os.Remove(absPath) }

	return absPath, cleanup, nil
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: fragletc [flags] <image>

Execute fraglet code in the specified container image.

Arguments:
  image         Container image to use (e.g., 100hellos/python:local)

Flags:
  -fraglet-path string
        Path where code is mounted in container (default: /FRAGLET)
  -file string
        Path to code file (if not provided, reads from STDIN)

Examples:
  # Read from STDIN with default fraglet-path
  echo "print('Hello, World!')" | fragletc 100hellos/python:local

  # Read from STDIN with custom fraglet-path
  echo "print('Hello!')" | fragletc -fraglet-path /custom/path 100hellos/python:local

  # Read from file with default fraglet-path
  fragletc -file hello.py 100hellos/python:local

  # Read from file with custom fraglet-path
  fragletc -fraglet-path /FRAGLET -file hello.py 100hellos/python:local

  # Pipe code
  cat hello.py | fragletc 100hellos/python:local

  # Short form with flags
  echo "console.log('test')" | fragletc -fraglet-path /FRAGLET 100hellos/javascript:local
`)
}
