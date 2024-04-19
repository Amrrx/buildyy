// cmd/cli/main.go
package main

import (
	"os"

	"github.com/spf13/cobra"
	"buildy/pkg/build"
	"buildy/pkg/config"
	"buildy/pkg/logging"
	"buildy/pkg/reporting"
)

var (
	configFile string
	outputDir  string
	logger     *logging.Logger
	rootCmd    = &cobra.Command{
		Use:   "build-automation-tool",
		Short: "A tool for automating builds, versioning, changelog, and tagging",
		Run:   runBuild,
	}
)

func init() {
	logger = logging.NewDefaultLogger()
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "build-config.yaml", "Path to the configuration file")
	rootCmd.PersistentFlags().StringVarP(&outputDir, "output", "o", "reports", "Output directory for build reports")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logger.Error.Println(err)
		os.Exit(1)
	}
}

func runBuild(cmd *cobra.Command, args []string) {
	// Parse the configuration file
	cfg, err := config.ParseConfig(configFile)
	if err != nil {
		logger.Error.Printf("Error parsing configuration file: %v\n", err)
		os.Exit(1)
	}

	// Run the build process
	buildResults := build.RunBuild(cfg, logger)

	// Generate and save the build report
	report, err := reporting.GenerateBuildReport(cfg, buildResults)
	if err != nil {
		logger.Error.Printf("Error generating build report: %v\n", err)
		os.Exit(1)
	}

	err = reporting.SaveBuildReport(report, outputDir)
	if err != nil {
		logger.Error.Printf("Error saving build report: %v\n", err)
		os.Exit(1)
	}

	logger.Info.Println("Build completed successfully")
}