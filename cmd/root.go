package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	debug          bool
	configFilePath string
)

func InitAndRunCommand() error {
	rootCmd := &cobra.Command{
		Use:   "root",
		Short: "Run the main process",
	}
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Run the main process",
		Run: func(cmd *cobra.Command, args []string) {
			if err := Run(); err != nil {
				fmt.Println(err)
				os.Exit(3)
			}
		},
	}
	rootCmd.AddCommand(startCmd)
	startCmd.PersistentFlags().BoolVar(&debug, "debug", false, "set log-level to debug")
	startCmd.PersistentFlags().StringVar(&configFilePath, "config-file", "config/config.yml", "set log-level to debug")
	return rootCmd.Execute()
}
