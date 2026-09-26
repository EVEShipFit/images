// Package atomicfile writes files so that an interrupted write is never
// mistaken for a complete one.
package atomicfile

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
)

// Copy writes everything r gives to filename, creating its directory.
func Copy(filename string, r io.Reader) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".write-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := io.Copy(tmp, r); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	// CreateTemp makes the file private; none of these files are secret.
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), filename)
}

func Write(filename string, raw []byte) error {
	return Copy(filename, bytes.NewReader(raw))
}
