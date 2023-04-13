package core

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	git "github.com/go-git/go-git/v5"
)

type SourceDir interface {
	Path() string
	AbsPath() string
	Validate() error
	Bootstrap() error
	Finalize(context Context, rollback bool, err error) error
	TemplateAbsPath() string
	HooksAbsPath() string
}

type LocalSourceDir struct {
	path string
}

func (l LocalSourceDir) Path() string {
	return l.path
}

func (l LocalSourceDir) AbsPath() string {
	// FIXME: what could go wrong?
	abs, _ := filepath.Abs(l.path)
	return abs
}

func (l LocalSourceDir) TemplateAbsPath() string {
	return filepath.Join(l.AbsPath(), "template")
}

func (l LocalSourceDir) HooksAbsPath() string {
	return filepath.Join(l.AbsPath(), "hooks")
}

// Validates wheather the path exists and is a directory
func (l LocalSourceDir) Validate() error {
	srcInfo, srcInfoErr := os.Stat(l.path)
	if srcInfoErr != nil {
		return srcInfoErr
	}
	if !srcInfo.IsDir() {
		return errors.New(fmt.Sprintf("%s is not a directory", l.path))
	}
	return nil
}

func (l LocalSourceDir) Bootstrap() error {
	return nil
}

func (l LocalSourceDir) Finalize(context Context, rollback bool, err error) error {
	return nil
}

// ---------------------------------------------------------------

type GitSourceDir struct {
	url      string
	destPath string
}

func (g GitSourceDir) Path() string {
	return g.destPath
}

func (g GitSourceDir) AbsPath() string {
	// FIXME: what could go wrong
	abs, _ := filepath.Abs(g.destPath)
	return abs
}

func (g GitSourceDir) TemplateAbsPath() string {
	return filepath.Join(g.AbsPath(), "template")
}

func (g GitSourceDir) HooksAbsPath() string {
	return filepath.Join(g.AbsPath(), "hooks")
}

// Validates wheather the path exists and is a directory
func (g GitSourceDir) Validate() error {
	return nil
}

func (g GitSourceDir) Bootstrap() error {
	if err := os.MkdirAll(g.destPath, fs.ModePerm); err != nil {
		return err
	}
	log.Printf("Cloning repository %s to %s", g.url, g.destPath)
	_, err := git.PlainClone(g.destPath, false, &git.CloneOptions{
		URL:          g.url,
		Progress:     os.Stdout,
		SingleBranch: true,
	})
	if err != nil {
		return err
	}
	return nil
}

func (g GitSourceDir) Finalize(context Context, rollback bool, err error) error {
	return nil
}
