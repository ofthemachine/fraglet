package fraglet

import (
	"fmt"
	"io"
	"os"

	"github.com/ofthemachine/fraglet/pkg/fraglet"
)

// Injector performs fraglet injection into a single target file.
type Injector struct{}

func NewInjector() *Injector {
	return &Injector{}
}

// Inject performs fraglet injection using the file injector.
// The injection config already contains the CodePath, so we just pass it through.
// If codePath is set without match/match_start markers, performs direct file replacement.
func (i *Injector) Inject(fragletPath string, injection fraglet.InjectionConfig) error {
	// Detect direct file replacement mode: codePath set, no match markers
	if injection.CodePath != "" && injection.Match == "" && injection.MatchStart == "" {
		if err := copyFile(fragletPath, injection.CodePath); err != nil {
			return fmt.Errorf("direct file replacement failed: %w", err)
		}
		// Remove temp fraglet file after copy
		_ = os.Remove(fragletPath)
		return nil
	}

	// Existing template injection logic
	if err := fraglet.InjectFile(fragletPath, &injection); err != nil {
		return err
	}
	// Remove temp fraglet file after injection
	_ = os.Remove(fragletPath)
	return nil
}

// copyFile copies the source file to the destination, preserving file permissions.
func copyFile(src, dst string) error {
	// Read source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Get source file info for permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	// Preserve destination file mode if it exists, otherwise use source mode
	mode := srcInfo.Mode()
	if dstInfo, err := os.Stat(dst); err == nil {
		mode = dstInfo.Mode()
	}

	// Create destination file
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}

	// Copy contents
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		dstFile.Close()
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Close file before chmod to ensure data is flushed
	if err := dstFile.Close(); err != nil {
		return fmt.Errorf("failed to close destination file: %w", err)
	}

	// Ensure file permissions are set correctly after writing
	if err := os.Chmod(dst, mode); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	return nil
}
