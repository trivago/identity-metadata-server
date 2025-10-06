package shared

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRotateSymlinkNoExist(t *testing.T) {
	assert := assert.New(t)
	tmpDir := t.TempDir()

	symlinkPath := filepath.Join(tmpDir, "testsymlink")
	testFile, err := os.CreateTemp(tmpDir, "testfile")
	assert.NoError(err)
	testFileName := testFile.Name()
	assert.NoError(testFile.Close())

	err = RotateSymlink(symlinkPath, testFileName)
	assert.NoError(err)

	linkStat, err := os.Lstat(symlinkPath)
	assert.NoError(err, "symlink should have been created")
	assert.Equal(linkStat.Mode()&os.ModeSymlink, os.ModeSymlink, "should be a symlink")

	testFilePath, err := ReadlinkAbs(symlinkPath)
	assert.NoError(err)
	assert.Equal(testFileName, testFilePath, "symlink should point to the test file")
}

func TestRotateSymlinkNoop(t *testing.T) {
	assert := assert.New(t)
	tmpDir := t.TempDir()

	symlinkPath := filepath.Join(tmpDir, "testsymlink")
	testFile, err := os.CreateTemp(tmpDir, "testfile")
	assert.NoError(err)
	testFileName := testFile.Name()
	assert.NoError(testFile.Close())

	err = os.Symlink(testFileName, symlinkPath)
	assert.NoError(err)

	err = RotateSymlink(symlinkPath, testFileName)
	assert.NoError(err)

	linkStat, err := os.Lstat(symlinkPath)
	assert.NoError(err)
	assert.Equal(linkStat.Mode()&os.ModeSymlink, os.ModeSymlink, "should be a symlink")

	testFilePath, err := ReadlinkAbs(symlinkPath)
	assert.NoError(err)
	assert.Equal(testFileName, testFilePath, "symlink should point to the test file")
}

func TestRotateSymlinkRotate(t *testing.T) {
	assert := assert.New(t)
	tmpDir := t.TempDir()

	symlinkPath := filepath.Join(tmpDir, "testsymlink")
	testFile, err := os.CreateTemp(tmpDir, "testfile")
	assert.NoError(err)
	testFileName := testFile.Name()
	assert.NoError(testFile.Close())

	testFile2, err := os.CreateTemp(tmpDir, "testfile")
	assert.NoError(err)
	testFile2Name := testFile2.Name()
	assert.NoError(testFile2.Close())

	err = os.Symlink(testFileName, symlinkPath)
	assert.NoError(err)

	err = RotateSymlink(symlinkPath, testFile2Name)
	assert.NoError(err)

	linkStat, err := os.Lstat(symlinkPath)
	assert.NoError(err)
	assert.Equal(linkStat.Mode()&os.ModeSymlink, os.ModeSymlink, "should be a symlink")

	testFilePath, err := ReadlinkAbs(symlinkPath)
	assert.NoError(err)
	assert.Equal(testFile2Name, testFilePath, "symlink should point to the second file")
}

func TestRotateSymlinkRotateConvert(t *testing.T) {
	assert := assert.New(t)
	tmpDir := t.TempDir()

	testFile, err := os.CreateTemp(tmpDir, "testfile")
	assert.NoError(err)
	testFileName := testFile.Name()
	assert.NoError(testFile.Close())

	testFile2, err := os.CreateTemp(tmpDir, "testfile")
	assert.NoError(err)
	testFile2Name := testFile2.Name()
	assert.NoError(testFile2.Close())

	symlinkPath := testFileName

	err = RotateSymlink(symlinkPath, testFile2Name)
	assert.NoError(err)

	linkStat, err := os.Lstat(symlinkPath)
	assert.NoError(err)
	assert.Equal(linkStat.Mode()&os.ModeSymlink, os.ModeSymlink, "should be a symlink")

	testFilePath, err := ReadlinkAbs(symlinkPath)
	assert.NoError(err)
	assert.Equal(testFile2Name, testFilePath, "symlink should point to the second file")

	generatedFiles, err := os.ReadDir(tmpDir)
	assert.NoError(err)
	assert.Equal(3, len(generatedFiles), "should have three files: the symlink and two test files")
}
