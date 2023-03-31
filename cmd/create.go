package cmd

import (
	"fmt"

	"github.com/jblawatt/gopier/core"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create from source",
	RunE: func(cmd *cobra.Command, args []string) error {
		src, _ := cmd.Flags().GetString("src")
		dest, _ := cmd.Flags().GetString("dest")
		values, _ := cmd.Flags().GetString("values")
		ctx := core.CreateDefaultContext(src, dest, values)
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

	rootCmd.AddCommand(createCmd)

}
