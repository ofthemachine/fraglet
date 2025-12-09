package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ofthemachine/fraglet/pkg/fraglet"
	"github.com/ofthemachine/fraglet/pkg/runner"
)

const defaultFragletPath = "/FRAGLET"

func main() {
	flag.Usage = usage

	// Define flags - use same variable for short and long forms
	image := flag.String("image", "", "Container image to use (e.g., 100hellos/python:local)")
	envelope := flag.String("envelope", "", "Use embedded envelope by name (e.g., python, javascript)")
	input := flag.String("input", "", "Path to code file (if not provided, reads from STDIN)")
	fragletPath := flag.String("fraglet-path", defaultFragletPath, "Path where code is mounted in container (default: /FRAGLET)")

	// Also define short forms that point to the same variables
	flag.StringVar(image, "i", "", "Container image (short form)")
	flag.StringVar(envelope, "e", "", "Use embedded envelope by name (short form)")
	flag.StringVar(input, "f", "", "Path to code file (short form)")
	flag.StringVar(fragletPath, "p", defaultFragletPath, "Path where code is mounted in container (short form)")

	flag.Parse()

	// Validate: must specify either image or envelope, but not both
	if *image == "" && *envelope == "" {
		usage()
		os.Exit(1)
	}
	if *image != "" && *envelope != "" {
		fmt.Fprintf(os.Stderr, "Error: cannot specify both --image and --envelope\n")
		os.Exit(1)
	}

	var code string
	var err error

	// Read code from file if provided, otherwise from STDIN
	if *input != "" {
		codeBytes, err := os.ReadFile(*input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", *input, err)
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

	// Determine container image and fraglet path
	var containerImage string
	var finalFragletPath string

	if *envelope != "" {
		// Use envelope (filesystem if FRAGLET_ENVELOPES_DIR is set, otherwise embedded)
		env, err := fraglet.NewFragletEnvironmentAuto()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading envelopes: %v\n", err)
			os.Exit(1)
		}

		envelopeObj, ok := env.GetRegistry().GetEnvelope(*envelope)
		if !ok {
			fmt.Fprintf(os.Stderr, "Error: envelope not found: %s\n", *envelope)
			os.Exit(1)
		}

		containerImage = envelopeObj.Container
		// Use envelope's fragletPath unless overridden
		if *fragletPath == defaultFragletPath {
			finalFragletPath = envelopeObj.FragletPath
		} else {
			finalFragletPath = *fragletPath
		}
	} else {
		// Use direct container image
		containerImage = *image
		finalFragletPath = *fragletPath
	}

	// Write code to temp file
	tmpFile, cleanup, err := writeTempFile(code)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temp file: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	// Create runner
	r := runner.NewRunner(containerImage, "")

	// Execute with volume mount at fragletPath
	spec := runner.RunSpec{
		Container: containerImage,
		Volumes: []runner.VolumeMount{
			{
				HostPath:      tmpFile,
				ContainerPath: finalFragletPath,
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
	fmt.Fprintf(os.Stderr, `Usage: fragletc [flags]

Execute fraglet code in a container using either --image or --envelope.

Flags:
  -e, --envelope string
        Use embedded envelope by name (e.g., python, javascript)
  -f, --input string
        Path to code file (if not provided, reads from STDIN)
  -i, --image string
        Container image to use (e.g., 100hellos/python:latest)
  -p, --fraglet-path string
        Path where code is mounted in container (default: /FRAGLET)

Examples:
  # Using container image
  echo 'print("Hello")' | fragletc --image 100hellos/python:latest
  fragletc --image 100hellos/python:latest --input script.py

  # Using embedded envelope
  echo 'print("Hello")' | fragletc --envelope python
  fragletc --envelope python --input script.py
`)
}
