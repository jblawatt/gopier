package cmd

import (
	"fmt"

	"github.com/jblawatt/gopier/core"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create from source",
	RunE: func(cmd *cobra.Command, args []string) error {
		// load options
		src, _ := cmd.Flags().GetString("src")
		dest, _ := cmd.Flags().GetString("dest")
		values, _ := cmd.Flags().GetString("values")
		dryRun, _ := cmd.Flags().GetBool("dryRun")

		options := &core.GopierOptions{
			DryRun:       dryRun,
			TemplatePath: viper.GetString(core.ConfigTemplateCache),
		}
		ctx, err := core.CreateDefaultContext(src, dest, values, options)
    if err != nil {
      return err
    }

		runner := core.CreateDefaultRunner()
		if err := runner.Run(ctx); err != nil {
			return fmt.Errorf("Error running template %w", err)
		}

		return nil
	},
}

func init() {

	createCmd.Flags().StringP("src", "s", "", "hello")
	createCmd.MarkFlagRequired("src")

	createCmd.Flags().StringP("dest", "d", "", "hello")
	createCmd.MarkFlagRequired("dest")

	createCmd.Flags().StringP("values", "v", "", "hello")

	createCmd.Flags().BoolP("dryRun", "", false, "run without creating destination or applying hooks")

	rootCmd.AddCommand(createCmd)

}
