package cmd

import (
	"log"

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
		ctx, ctxErr := core.CreateContext(values)
		if ctxErr != nil {
			return ctxErr
		}
		log.Println(ctx)
		if err := core.CopyNew(src, dest, ctx); err != nil {
			return err
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
