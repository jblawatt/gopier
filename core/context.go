package core

import (
	"path/filepath"
	"strings"

	"github.com/gosimple/slug"
	"github.com/spf13/viper"
)

type Context interface {
	Src() SourceDir
	Dest() DestDir
	Options() *GopierOptions
	TemplateContext() *TemplateContext
}

type DefaultContext struct {
	src             SourceDir
	dest            DestDir
	templateContext *TemplateContext
	options         *GopierOptions
}

func CreateSourceDir(src string) SourceDir {
	if strings.HasPrefix(src, GitPrefix) {
		src = strings.TrimPrefix(src, GitPrefix)
		slug_ := slug.Make(src)
		templatesCacheDir := viper.GetString(ConfigTemplateCache)
		return GitSourceDir{
			url:      src,
			destPath: filepath.Join(templatesCacheDir, slug_),
		}
	}
	return LocalSourceDir{src}
}

func CreateDefaultContext(src string, dest string, valuesFile string, options *GopierOptions) (Context, error) {
	sourceDir := CreateSourceDir(src)

	vf, _ := filepath.Abs(valuesFile)
	templateContext, terr := CreateTemplateContext(
		filepath.Join(sourceDir.AbsPath(), "values.yaml"),
		vf,
	)
	if terr != nil {
		return nil, terr
	}
	return &DefaultContext{
		sourceDir,
		LocalDestDir{dest},
		templateContext,
		options,
	}, nil
}

func (d *DefaultContext) Src() SourceDir {
	return d.src
}

func (d *DefaultContext) Dest() DestDir {
	return d.dest
}

func (d *DefaultContext) Options() *GopierOptions {
	return d.options
}

func (d *DefaultContext) TemplateContext() *TemplateContext {
	return d.templateContext
}
