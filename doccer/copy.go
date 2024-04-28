package doccer

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

func CopyDirectory(fileSys fs.FS, scrDir, dest string) error {
	entries, err := fs.ReadDir(fileSys, scrDir)
	if err != nil {
		return errors.Wrapf(err, "failed to read directory: '%s'", scrDir)
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(scrDir, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		fileInfo, err := fs.Stat(fileSys, sourcePath)
		if err != nil {
			return errors.Wrapf(err, "failed to get file info: '%s (%s)'", sourcePath, entry.Name())
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := CreateIfNotExists(destPath, 0755); err != nil {
				return err
			}
			if err := CopyDirectory(fileSys, sourcePath, destPath); err != nil {
				return err
			}
		case os.ModeSymlink:
		default:
			if err := Copy(fileSys, sourcePath, destPath); err != nil {
				return err
			}
		}

		fInfo, err := entry.Info()
		if err != nil {
			return errors.Wrapf(err, "failed to get file info: '%s'", sourcePath)
		}

		isSymlink := fInfo.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			if err := os.Chmod(destPath, fInfo.Mode()); err != nil {
				return errors.Wrapf(err, "failed to change mode of file: '%s'", destPath)
			}
		}
	}
	return nil
}

func Copy(filesys fs.FS, srcFile, dstFile string) error {
	out, err := os.Create(dstFile)
	if err != nil {
		return errors.Wrapf(err, "failed to create file: '%s'", dstFile)
	}

	defer out.Close()

	in, err := filesys.Open(srcFile)
	if err != nil {
		return errors.Wrapf(err, "failed to open file: '%s'", srcFile)
	}

	defer in.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return errors.Wrapf(err, "failed to copy file: '%s' to '%s'", srcFile, dstFile)
	}

	return nil
}

func Exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}

func CreateIfNotExists(dir string, perm os.FileMode) error {
	if Exists(dir) {
		return nil
	}

	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}

	return nil
}
