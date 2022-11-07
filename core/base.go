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
	"path/filepath"

	yaml "gopkg.in/yaml.v3"
)

func verifySrc(src string) error {
	srcInfo, srcInfoErr := os.Stat(src)
	if srcInfoErr != nil {
		return srcInfoErr
	}
	if !srcInfo.IsDir() {
		return errors.New(fmt.Sprintf("%s is not a directory", src))
	}
	return nil
}

func verifyDest(dest string) error {
	_, destInfoErr := os.Stat(dest)
	if destInfoErr != nil && errors.Is(destInfoErr, os.ErrNotExist) {
		os.MkdirAll(dest, fs.ModePerm)
		return nil
	}
	if dirContent, err := os.ReadDir(dest); err != nil {
		return err
	} else {
		if len(dirContent) != 0 {
			return errors.New(fmt.Sprintf("%s is not empty", dest))
		}
	}
	return nil
}

func CopyNew(src string, dest string, ctx TemplateContext) error {
	if err := verifySrc(src); err != nil {
		return err
	}
	if err := verifyDest(dest); err != nil {
		return err
	}
	return handleFolder(src, dest, ctx)
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

func handleFolder(src string, dest string, ctx TemplateContext) error {
	items, _ := os.ReadDir(src)
	for _, item := range items {
		if item.IsDir() {
			itemName := interpolate(item.Name(), ctx)
			destName := interpolate(dest, ctx)
			os.Mkdir(filepath.Join(destName, itemName), fs.ModePerm)
			err := handleFolder(
				filepath.Join(src, item.Name()),
				filepath.Join(dest, item.Name()),
				ctx,
			)
			if err != nil {
				return err
			}
		} else {
			itemName := interpolate(item.Name(), ctx)
			dest = interpolate(dest, ctx)
			destName := filepath.Join(dest, itemName)
			srcFile, _ := os.ReadFile(filepath.Join(src, item.Name()))
			if filepath.Ext(item.Name()) == TemplateExt {
				destName = destName[:len(destName)-len(TemplateExt)]
				tmpl, _ := template.New("").Parse(string(srcFile))
				out, _ := os.Create(destName)
				defer out.Close()
				tmpl.Execute(out, ctx)
			} else {
				os.WriteFile(destName, srcFile, 0666)

			}

		}
	}
	return nil
}

func CreateContext(filename string) (TemplateContext, error) {
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
