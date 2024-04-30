// cmd/cli/main.go
package main

import (
	"os"

	"buildy/pkg/build"
	"buildy/pkg/config"
	"buildy/pkg/logging"
	"buildy/pkg/reporting"
	"buildy/pkg/changelog"
	"buildy/pkg/versioning"
	"github.com/spf13/cobra"
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

	for i, subProject := range cfg.SubProjects {
		if buildResults[subProject.Name] == nil {

			versionIncrement := "patch"
			
			// Increment the version if the build was successful
			newVersion := versioning.IncrementVersion(subProject.Version, versionIncrement)
			cfg.SubProjects[i].Version = newVersion
			logger.Info.Printf("Subproject %s version updated to %s\n", subProject.Name, newVersion)
		}
	}

	//Increment central version
	newCentralVersion := versioning.IncrementVersion(cfg.Version, "patch")
	cfg.Version = newCentralVersion
	logger.Info.Printf("Central Project %s version updated to %s\n", cfg.Name, newCentralVersion)


	// Generate the changelog
	err = changelog.GenerateChangelogs(cfg, outputDir)
	if err != nil {
		logger.Error.Printf("Error generating changelog: %v\n", err)
		os.Exit(1)
	}

	// Save the updated configuration file
	err = config.SaveConfig(configFile, cfg)
	if err != nil {
		logger.Error.Printf("Error saving configuration file: %v\n", err)
		os.Exit(1)
	}

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


func getSubProjectByName(subProjects []config.SubProject, name string) *config.SubProject {
	for _, subProject := range subProjects {
		if subProject.Name == name {
			return &subProject
		}
	}
	return nil
}


// TODO: Fix the increment of the central version as well [It doesn't increment] - Done
// TODO: Fix the replicating of central history log [It keepy re-writing the whole history for central] - Done
// TODO: A mechanism to stop writing logs if no new commits, maybe increase the version or introduce a build number
