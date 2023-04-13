package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
  Use: "config",
  Short: "show current config",
  RunE: runConfigCmd,
}

func runConfigCmd(cmd *cobra.Command, args []string) error {
  fmt.Printf("Using config file %s\n\n", viper.GetViper().ConfigFileUsed())

  for _, key := range viper.AllKeys() {
    fmt.Printf("%s = %s\n", key, viper.Get(key))
  }

  return nil
}

func init() {
  rootCmd.AddCommand(configCmd)
}
