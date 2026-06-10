package atomicfile

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// WriteFile atomically writes the contents of r to the specified filepath with the given mode.
// This is a copy of https://github.com/natefinch/atomic/blob/master/atomic.go with minor modifications allowing
// to set mode of the written file. If the file already exists, its mode is preserved.
func WriteFile(filename string, r io.Reader, mode os.FileMode) (err error) {
	// write to a temp file first, then we'll atomically replace the target file
	// with the temp file.
	dir, file := filepath.Split(filename)
	if dir == "" {
		dir = "."
	}

	f, err := os.CreateTemp(dir, file)
	if err != nil {
		return fmt.Errorf("cannot create temp file: %w", err)
	}
	defer func() {
		if err != nil {
			// Don't leave the temp file lying around on error.
			_ = os.Remove(f.Name()) // yes, ignore the error, not much we can do about it.
		}
	}()
	// ensure we always close f. Note that this does not conflict with the close below, as close is idempotent while
	// it returns an error for repeating close operations.
	defer f.Close() //nolint:errcheck
	name := f.Name()
	if _, err = io.Copy(f, r); err != nil {
		return fmt.Errorf("cannot write data to tempfile %q: %w", name, err)
	}
	// fsync is important, otherwise os.Rename could rename a zero-length file
	if err = f.Sync(); err != nil {
		return fmt.Errorf("can't flush tempfile %q: %w", name, err)
	}
	if err = f.Close(); err != nil {
		return fmt.Errorf("can't close tempfile %q: %w", name, err)
	}

	// get the file mode from the original file and use that for the replacement file, too.
	destInfo, err := os.Stat(filename)
	if os.IsNotExist(err) {
		// no original file
		if err = os.Chmod(name, mode); err != nil {
			return fmt.Errorf("can't set filemode on tempfile %q: %w", name, err)
		}
	} else if err != nil {
		return err
	} else {
		sourceInfo, err := os.Stat(name)
		if err != nil {
			return err
		}

		if sourceInfo.Mode() != destInfo.Mode() {
			if err = os.Chmod(name, destInfo.Mode()); err != nil {
				return fmt.Errorf("can't set filemode on tempfile %q: %w", name, err)
			}
		}
	}
	if err := os.Rename(name, filename); err != nil {
		return fmt.Errorf("cannot replace %q with tempfile %q: %w", filename, name, err)
	}
	return nil
}
