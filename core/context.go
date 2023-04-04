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
	Options() Options
	TemplateContext() TemplateContext
}

type DefaultContext struct {
	src        SourceDir
	dest       DestDir
	valuesFile string
	options    Options
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

func CreateDefaultContext(src string, dest string, valuesFile string) Context {
	vf, _ := filepath.Abs(valuesFile)
	sourceDir := CreateSourceDir(src)
	return DefaultContext{
		sourceDir,
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
