package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	debug          bool
	configFilePath string
	logLevel string
	logFormat string
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
	startCmd.PersistentFlags().StringVar(&configFilePath, "config-file", "config/config.yml", "set log-level to debug")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "v", "info", "Logger log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "text", "Logger logs format (text, json)")
	return rootCmd.Execute()
}
