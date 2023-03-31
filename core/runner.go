package core

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"os"
	"path/filepath"
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

	items, _ := handleItem(
		context.Src().TemplateAbsPath(),
		context.Dest().AbsPath(),
		context,
	)

	if !context.Options().DryRun {
		writeItems(items)
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

func validateName(itemName string) error {
	if itemName == "" {
		return fmt.Errorf("Item name sould not be empty.")
	}
	return nil
}

func handleItem(src string, dest string, context Context) ([]RenderItem, error) {
	var out []RenderItem
	items, err := os.ReadDir(src)
	if err != nil {
		return nil, fmt.Errorf("Error reading template (sub)dir: %w", err)
	}

	for _, item := range items {
		if item.IsDir() {
			itemName := interpolate(item.Name(), context.TemplateContext())
			if err := validateName(itemName); err != nil {
				return nil, err
			}
			destName := interpolate(dest, context.TemplateContext())
			renderItem := RenderItem{
				Dir:     destName,
				Name:    itemName,
				Content: []byte{},
				Append:  false,
				IsDir:   true,
			}
			log.Printf("----- DIRECTORY: %s::%s\n", destName, itemName)
			out = append(out, renderItem)
			subitems, _ := handleItem(
				filepath.Join(src, item.Name()),
				filepath.Join(destName, itemName),
				context)
			out = append(out, subitems...)
		} else {
			destName := interpolate(dest, context.TemplateContext())
			itemName := interpolate(item.Name(), context.TemplateContext())

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

			// TODO: stehengeblieben

			log.Printf("----- FILE: %s\n", filepath.Join(destName, itemName))
			log.Println(string(srcContent))

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
