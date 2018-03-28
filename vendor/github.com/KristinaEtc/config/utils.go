package config

import (
	"os"
	"path/filepath"

	"github.com/kardianos/osext"
)

func GetPathForDir(path string) (string, error) {

	if filepath.IsAbs(path) == true {
		return path, nil
	}
	filename, err := osext.Executable()
	if err != nil {
		return "", err
	}

	fpath := filepath.Dir(filename)
	fpath = filepath.Join(fpath, path)
	return fpath, nil

}

// Exists returns whether the given file or directory exists or not.
func Exists(path string) (bool, error) {

	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
