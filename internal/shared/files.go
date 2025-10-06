package shared

import (
	"os"
	"path/filepath"
	"strconv"
)

// ReadlinkAbs reads the target of a symlink and returns its absolute path.
// If the symlink does not exist, or is not a symlink, it returns an error.
// If the symlink points to a relative path, it resolves it relative to the symlink's directory.
// If the symlink points to an absolute path, it returns that path as is.
func ReadlinkAbs(path string) (string, error) {
	linkPath, err := os.Readlink(path)
	if err != nil {
		return "", WrapErrorf(err, "failed to resolve symlink %s", path)
	}

	// If the link is absolute, return it as is
	if filepath.IsAbs(linkPath) {
		return filepath.Clean(linkPath), nil
	}

	// If the link is relative, resolve it relative to the symlink's directory
	symlinkDir := filepath.Dir(path)
	absolutePath, err := filepath.Abs(filepath.Join(symlinkDir, linkPath))
	if err != nil {
		return "", WrapErrorf(err, "failed to resolve absolute path for symlink %s", path)
	}
	return absolutePath, nil
}

// RotateSymlink points an existing symlink to a new target path.
// It is a convenience function that uses an undo stack to allow rolling back changes.
func RotateSymlink(symlinkPath, targetPath string) error {
	undo := NewUndoStack()
	err := RotateSymlinkWithUndo(symlinkPath, targetPath, undo)
	return undo.RollbackIfError(err)
}

// RotateSymlinkList rotates a list of symlinks to their respective target paths.
// SymlinkPath is the key and TargetPath is the value in the KVList.
// It uses an undo stack to allow rolling back changes if any of the rotations fails.
func RotateSymlinkList(symlinkToTarget *KVList[string, string]) error {
	undoStack := NewUndoStack()
	var err error

	symlinkToTarget.ForEach(func(symlinkPath, targetPath string) bool {
		err = RotateSymlinkWithUndo(symlinkPath, targetPath, undoStack)
		// Return true to continue iterating, false to stop ("is OK")
		// This causes the first error to stop the iteration, exposing
		// the first error found to the caller.
		return err == nil
	})

	return undoStack.RollbackIfError(err)
}

// RotateSymlinkWithUndo points an existing symlink to a new target path.
// The operation is undo-able via the provided stack; the function itself
// returns only an error – success is indicated by a nil error.
// If the symlink already points to the target, it does nothing.
// If the symlink does not exist, it creates a new symlink.
// If symlinkPath is a regular file, it renames it to a timestamped
// version before creating the symlink. The timestamp format is YYYYMMDDHHMMSS
// and based on the last modification time of the file.
// An undo stack is used to allow rolling back the changes if needed.
func RotateSymlinkWithUndo(symlinkPath, targetPath string, undoStack *UndoStack) error {
	var previousPath string

	// Check if the target path exists
	if _, err := os.Stat(targetPath); err != nil {
		return WrapErrorf(err, "failed to get file info for target %s", targetPath)
	}

	// Make sure to use Lstat() to check if the file is a symlink
	// as Stat() would follow the symlink
	fileInfo, err := os.Lstat(symlinkPath)
	if err != nil && !os.IsNotExist(err) {
		return WrapErrorf(err, "failed to get file info for symlink %s", symlinkPath)
	}

	switch {
	case os.IsNotExist(err):
		// File does not exist, just create a new symlink (do nothing)

	case fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink:
		// File is a symlink
		// Check if the symlink points to the same target
		previousPath, err = ReadlinkAbs(symlinkPath)
		if err != nil {
			return WrapErrorf(err, "failed to read symlink %s", symlinkPath)
		}
		absTargetPath, resolveErr := filepath.Abs(targetPath)

		// We check resolveErr to avoid false positives like when comparing two
		// empty strings. Any error in filepath.Abs is ignored here, as it will
		// either surface later or not affect the symlink creation.
		if previousPath == absTargetPath && resolveErr == nil {
			// No need to rotate, the symlink already points to the target
			return nil
		}

		// Get the exact value of the symlink for rollback
		rollbackPath, err := os.Readlink(symlinkPath)
		if err != nil {
			return WrapErrorf(err, "failed to read symlink %s", symlinkPath)
		}

		// Remove the existing symlink
		if err := os.Remove(symlinkPath); err != nil {
			return WrapErrorf(err, "failed to remove existing symlink %s -> %s", symlinkPath, previousPath)
		}
		undoStack.Push(func() error { return os.Symlink(rollbackPath, symlinkPath) })

	default:
		// File is not a symlink, rename it to a timestamped version.
		suffix := fileInfo.ModTime().Format("20060102150405")
		previousPath = symlinkPath + "." + suffix

		// Check if file exists and append an increasing number suffix until
		// we find a unique name. This is a O(n) operation but it is unlikely
		// to be a problem in practice (unless the symlink is rotated multiple
		// times per second).
		for i := 1; ; i++ {
			if _, err := os.Stat(previousPath); os.IsNotExist(err) {
				break
			}
			previousPath = symlinkPath + "." + suffix + "-" + strconv.Itoa(i)
		}

		if err = os.Rename(symlinkPath, previousPath); err != nil {
			return WrapErrorf(err, "failed to rename %s to %s", symlinkPath, previousPath)
		}
		undoStack.Push(func() error { return os.Rename(previousPath, symlinkPath) })
	}

	symlinkDir := filepath.Dir(symlinkPath)
	relativeTargetPath, err := filepath.Rel(symlinkDir, targetPath)
	if err != nil {
		return undoStack.RollbackFromError(WrapErrorf(err, "failed to get relative path from %s to %s", symlinkDir, targetPath))
	}

	// Create a new symlink
	if err := os.Symlink(relativeTargetPath, symlinkPath); err != nil {
		return undoStack.RollbackFromError(WrapErrorf(err, "failed to create symlink %s -> %s", symlinkPath, targetPath))
	}

	return nil
}
