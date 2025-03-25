package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ng-spec",
	Short: "A simple CLI tool to generate minimal integration test boilerplate for Angular components",
	Long:  `Generate minimal integration test boilerplate for Angular components using the Angular Testing Library.`,
	Example: `
	ng-spec
	ng-spec app
	ng-spec /path/to/component
	`,
	Run: func(cmd *cobra.Command, args []string) {
		var component string

		fmt.Println(component)

		if len(args) > 0 {
			component = args[0]
		}

		generateComponentTest(component)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
