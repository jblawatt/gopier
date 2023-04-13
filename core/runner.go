package core

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Runner interface {
	Run(context Context) error
}

type DefaultRunner struct{}

func formatError(msg string, err error) error {
	return fmt.Errorf(msg, err)
}

func (p *DefaultRunner) Run(context Context) error {

	if err := context.Src().Validate(); err != nil {
		return formatError("Error validating source: %w", err)
	}

	if err := context.Dest().Validate(); err != nil {
		return formatError("Error validating destination: %w", err)
	}

	if err := context.Src().Bootstrap(); err != nil {
		return formatError("Error bootstrapping source: %w", err)
	}

	// Render source to destination structure
	items, err := p.renderSourceTemplates(
		context.Src().TemplateAbsPath(),
		context.Dest().AbsPath(),
		context,
	)

	if err != nil {
		return formatError("Error rendering source template: %w", err)
	}

	// Write Items to files
	if !context.Options().DryRun {
		if err := p.writeToDestination(items); err != nil {
			if ferr := p.finalize(context, true, err); ferr != nil {
				return fmt.Errorf("Error writing to destination: %w\nError cleaning up during rollback: %w", err, ferr)
			}
			return formatError("Error writing to destination: %w", err)
		}
	}

	// Apply Hooks
	if !context.Options().DryRun {
    if err := p.applyHooks(context); err != nil {
			if ferr := p.finalize(context, true, err); ferr != nil {
				return fmt.Errorf("Error applying hooks to destination: %w\nError cleaning up during rollback: %w", err, ferr)
			}
			return formatError("Error applying hooks destination: %w", err)
    }
	}

	context.Src().Finalize(context, false, nil)

	return nil
}

func (p *DefaultRunner) writeToDestination(items []RenderItem) error {
	for _, item := range items {
		if item.IsDir {
			os.MkdirAll(filepath.Join(item.Dir, item.Name), fs.ModePerm)
		} else {
			os.WriteFile(filepath.Join(item.Dir, item.Name), item.Content, fs.ModePerm)
		}
	}
	return nil
}

func (p *DefaultRunner) finalize(context Context, rollback bool, err error) error {
	if err := context.Src().Finalize(context, rollback, err); err != nil {
		return formatError("Error finalizing source: %w", err)
	}
	if rollback {
		if err := os.RemoveAll(context.Dest().AbsPath()); err != nil {
			return formatError("Error cleanup destination while rollback: %w", err)
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
	if !strings.HasPrefix(destPath, context.Dest().AbsPath()) {
		return fmt.Errorf("Dest path escapes from destination")
	}
	return nil
}

func (p *DefaultRunner) renderSourceTemplates(src string, dest string, context Context) ([]RenderItem, error) {
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
			if subitems, err := p.renderSourceTemplates(subSrc, subDest, context); err != nil {
				return nil, err
			} else {
				out = append(out, subitems...)
			}
		} else {
			// if is template
			var srcContent []byte
      tplExt := context.Options().TemplateExt
			if filepath.Ext(item.Name()) == tplExt {
				templateContent, _ := os.ReadFile(filepath.Join(src, item.Name()))
				itemName = strings.TrimSuffix(itemName, tplExt)
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

func (p *DefaultRunner) applyHooks(context Context) error {
	log.Println(">>>> appling hooks")
	defer log.Println("<<<< hooks applied")
	srcHooksPath := context.Src().HooksAbsPath()
	if _, err := os.Stat(srcHooksPath); os.IsNotExist(err) {
		log.Println("=== no hooks")
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
		cmd := exec.Command(hookFile, context.Dest().AbsPath())
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
