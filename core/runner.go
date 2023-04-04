package core

import (
	"bytes"
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

type Runner interface {
	Run(context Context) error
}

type DefaultRunner struct{}

func (p *DefaultRunner) Run(context Context) error {
	if err := context.Src().Validate(); err != nil {
		return fmt.Errorf("Error validating source: %w", err)
	}

	if err := context.Dest().Validate(); err != nil {
		return fmt.Errorf("Error validating destination: %w", err)
	}

	context.Src().Bootstrap()

	// Create Items
	items, _ := handleItem(
		context.Src().TemplateAbsPath(),
		context.Dest().AbsPath(),
		context,
	)

	// Write Items to files
	if !context.Options().DryRun {
		writeItems(items)
	}

	// Apply Hooks
	if !context.Options().DryRun {
		applyHooks(context)
	}

	context.Src().Finalize()

	return nil
}

func writeItems(items []RenderItem) error {
	for _, item := range items {
		if item.IsDir {
			os.MkdirAll(filepath.Join(item.Dir, item.Name), fs.ModePerm)
		} else {
			// from original
			os.WriteFile(filepath.Join(item.Dir, item.Name), item.Content, fs.ModePerm)
		}
	}
	return nil
}

func validateName(context Context, itemName string) error {
	if itemName == "" {
		return fmt.Errorf("Item name sould not be empty.")
	}
	return nil
}

func validateDestPath(context Context, destPath string) error {
	// TODO: do not excape from dest
	return nil
}

func handleItem(src string, dest string, context Context) ([]RenderItem, error) {
	var out []RenderItem
	items, err := os.ReadDir(src)
	if err != nil {
		return nil, fmt.Errorf("Error reading template (sub)dir: %w", err)
	}

	for _, item := range items {
		itemName := interpolate(item.Name(), context.TemplateContext())
		if err := validateName(context, itemName); err != nil {
			return nil, err
		}
		destName := interpolate(dest, context.TemplateContext())
		if err := validateDestPath(context, destName); err != nil {
			return nil, err
		}

		if item.IsDir() {
			// add current item
			renderItem := RenderItem{
				Dir:     destName,
				Name:    itemName,
				Content: []byte{},
				Append:  false,
				IsDir:   true,
			}
			out = append(out, renderItem)

			// append subitems
			subSrc := filepath.Join(src, item.Name())
			subDest := filepath.Join(destName, itemName)
			if subitems, err := handleItem(subSrc, subDest, context); err != nil {
				return nil, err
			} else {
				out = append(out, subitems...)
			}
		} else {

			// if is template
			var srcContent []byte
			if filepath.Ext(item.Name()) == TemplateExt {
				// strip TemplateExtension
				itemName = itemName[:len(itemName)-len(TemplateExt)]
				templateContent, _ := os.ReadFile(filepath.Join(src, item.Name()))
				tmpl, _ := template.New("").Parse(string(templateContent))
				var w bytes.Buffer
				tmpl.Execute(&w, context.TemplateContext())
				srcContent = []byte(w.String())
			} else {
				srcContent, _ = os.ReadFile(filepath.Join(src, item.Name()))
			}

			renderItem := RenderItem{
				Dir:     destName,
				Name:    itemName,
				Content: srcContent,
				Append:  false,
				IsDir:   false,
			}

			out = append(out, renderItem)
		}

	}
	return out, nil
}

func CreateDefaultRunner() DefaultRunner {
	return DefaultRunner{}
}

type RenderItem struct {
	Dir     string
	Name    string
	Content []byte
	Append  bool
	IsDir   bool
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
