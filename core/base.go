package core

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Options struct {
	DryRun       bool
	TemplatePath string
}

type ValuesMap = map[interface{}]interface{}

func init() {

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("GOPIER")
	viper.AddConfigPath("$HOME/.gopier")

	// FIXME: $HOME not interpolating
	viper.SetDefault(ConfigTemplateCache, "/home/j3nko/.gopier/templates")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Errorf("Error loading config %w", err)
		}
	} else {
		log.Println("Configfile loaded")
	}
}
