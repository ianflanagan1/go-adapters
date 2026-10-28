// Package files adapts a directory of the filesystem. Paths are relative to
// the directory and can't escape it.
package files

import (
	"fmt"
	"os"
)

// Dir is a directory, opened with Open.
type Dir struct {
	root *os.Root
}

// Open opens the directory at path.
func Open(path string) (*Dir, error) {
	root, err := os.OpenRoot(path)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}
	return &Dir{root: root}, nil
}

// ReadFile returns the contents of the file at name.
func (d *Dir) ReadFile(name string) ([]byte, error) {
	content, err := d.root.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", name, err)
	}
	return content, nil
}

// WriteFile replaces the file at name with content, creating it if needed.
func (d *Dir) WriteFile(name string, content []byte) error {
	if err := d.root.WriteFile(name, content, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", name, err)
	}
	return nil
}

// Close releases the directory.
func (d *Dir) Close() error {
	if err := d.root.Close(); err != nil {
		return fmt.Errorf("closing the directory: %w", err)
	}
	return nil
}
