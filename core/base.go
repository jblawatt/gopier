package core

import (
	"path/filepath"
	"log"
	"os"

	"github.com/spf13/viper"
)

type ValuesMap = map[interface{}]interface{}

func init() {

  home, _ := os.UserHomeDir()
  viper.AddConfigPath(home)
  viper.AddConfigPath(".")
  viper.SetConfigType("yaml")
  viper.SetConfigFile(".gopier")
  viper.SetEnvPrefix("GOPIER_")

  cacheDir, _ := os.UserCacheDir()
  viper.SetDefault(ConfigTemplateCache, filepath.Join(cacheDir, "gopier", "templates"))
  viper.SetDefault(ConfigTemplateExt, ".tpl")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("Error loading config %s\n", err.Error())
		}
	} else {
		log.Println("Configfile loaded")
	}

}
