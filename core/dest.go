package core

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type DestDir interface {
	Path() string
	AbsPath() string
	Validate() error
}

type LocalDestDir struct {
	path string
}

func (l LocalDestDir) Path() string {
	return l.path
}

func (l LocalDestDir) AbsPath() string {
	abs, _ := filepath.Abs(l.path)
	return abs
}

func (l LocalDestDir) Validate() error {
	_, destInfoErr := os.Stat(l.path)
	if destInfoErr != nil && errors.Is(destInfoErr, os.ErrNotExist) {
		os.MkdirAll(l.path, fs.ModePerm)
		return nil
	}
	if dirContent, err := os.ReadDir(l.path); err != nil {
		return err
	} else {
		if len(dirContent) != 0 {
			return errors.New(fmt.Sprintf("%s is not empty", l.path))
		}
	}
	return nil
}
