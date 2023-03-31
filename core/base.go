package core

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

type SourceDir interface {
	Path() string
	AbsPath() string
	Validate() error
	Bootstrap() error
	Finalize() error
	TemplateAbsPath() string
	HooksAbsPath() string
}

type DestDir interface {
	Path() string
	AbsPath() string
	Validate() error
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

func (l LocalSourceDir) Finalize() error {
	return nil
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

type Options struct {
	DryRun bool
}

type Context interface {
	Src() SourceDir
	Dest() DestDir
	Options() Options
	TemplateContext() TemplateContext
}

type ValuesMap = map[interface{}]interface{}

type DefaultContext struct {
	src        SourceDir
	dest       DestDir
	valuesFile string
	options    Options
}

func CreateDefaultContext(src string, dest string, valuesFile string) Context {
	vf, _ := filepath.Abs(valuesFile)
	return DefaultContext{
		LocalSourceDir{src},
		LocalDestDir{dest},
		// FIXME: merge multiple values
		vf,
		Options{DryRun: false},
	}
}

func (d DefaultContext) Src() SourceDir {
	return d.src
}

func (d DefaultContext) Dest() DestDir {
	return d.dest
}

func (d DefaultContext) Options() Options {
	return d.options
}

func (d DefaultContext) TemplateContext() TemplateContext {
	ctx, err := CreateTemplateContext(d.valuesFile)
	if err != nil {
		panic(err)
	}
	return ctx
}

func applyHooks(ctx Context) error {
	srcHooksPath := ctx.Src().HooksAbsPath()
	if _, err := os.Stat(srcHooksPath); os.IsNotExist(err) {
		return nil
	}
	items, _ := os.ReadDir(srcHooksPath)
	for _, item := range items {
		if item.IsDir() {
			continue
		}
		hookFile := filepath.Join(srcHooksPath, item.Name())
		// FIXME: only run if executable
		// TODO: more paramters?
		cmd := exec.Command(hookFile, ctx.Dest().AbsPath())
		if output, err := cmd.Output(); err != nil {
			return err
		} else {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				log.Printf("%s: %s\n", item.Name(), line)
			}
		}
	}
	return nil
}

const TemplateExt = ".tpl"

type TemplateContext struct {
	Values map[interface{}]interface{}
}

func interpolate(strOrTemplate string, ctx TemplateContext) string {
	tpl, _ := template.New("").Parse(strOrTemplate)
	buff := new(bytes.Buffer)
	tpl.Execute(buff, ctx)
	processed := buff.String()
	log.Printf("interpolating string %s to %s\n", strOrTemplate, processed)
	return processed
}

func CreateTemplateContext(filename string) (TemplateContext, error) {
	content, rerr := ioutil.ReadFile(filename)
	if rerr != nil {
		return TemplateContext{}, fmt.Errorf("Error reading values file %s: %w", filename, rerr)
	}
	values := make(map[interface{}]interface{})
	merr := yaml.Unmarshal(content, &values)
	if merr != nil {
		return TemplateContext{}, fmt.Errorf("Error unmarshalling values file: %w", merr)
	}

	return TemplateContext{Values: values}, nil
}
