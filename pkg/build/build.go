// pkg/build/build.go
package build

import (

	"buildy/pkg/config"
	"buildy/pkg/logging"
)

func RunBuild(cfg *config.Config, logger *logging.Logger) map[string]error {
	buildResults := make(map[string]error)

	for _, subProject := range cfg.SubProjects {
		logger.Info.Printf("Building subproject: %s\n", subProject.Name)

		// TODO: Implement actual build steps for each subproject
		// For now, we'll just simulate a successful build
		buildResults[subProject.Name] = nil

		logger.Info.Printf("Subproject %s built successfully\n", subProject.Name)
	}

	return buildResults
}