// pkg/analyzer/analyzer.go
package analyzer

import (
	"strings"

	"build-automation-tool/pkg/config"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func DetermineVersionIncrement(subProject config.SubProject, repo *git.Repository) string {
	// Approach 1: Analyze commit messages
	increment := analyzeCommitMessages(subProject, repo)
	if increment != "" {
		return increment
	}

	// Approach 2: Check for specific files or patterns
	increment = checkSpecificFiles(subProject)
	if increment != "" {
		return increment
	}

	// Approach 3: Use a default increment (e.g., patch)
	return "patch"
}

func analyzeCommitMessages(subProject config.SubProject, repo *git.Repository) string {
	// Retrieve the commits for the subproject
	commits, err := getCommitsForSubproject(subProject, repo)
	if err != nil {
		return ""
	}

	// Analyze the commit messages to determine the version increment
	for _, commit := range commits {
		message := commit.Message
		if strings.Contains(message, "BREAKING CHANGE") || strings.Contains(message, "!:") {
			return "major"
		}
		if strings.Contains(message, "feat:") || strings.Contains(message, "feature:") {
			return "minor"
		}
	}

	return ""
}

func checkSpecificFiles(subProject config.SubProject) string {
	// Check for specific files or patterns in the subproject
	// For example, you can check for the presence of a "BREAKING_CHANGES" file
	// or analyze the diff of specific files to determine the version increment

	// Placeholder logic, modify as per your requirements
	if subProject.HasBreakingChanges {
		return "major"
	} else if subProject.HasNewFeatures {
		return "minor"
	}

	return ""
}

func getCommitsForSubproject(subProject config.SubProject, repo *git.Repository) ([]*object.Commit, error) {
	// Retrieve the commits for the subproject based on the repository and subproject path
	// Implement the logic to fetch the relevant commits for the subproject

	// Placeholder logic, modify as per your requirements
	commits := []*object.Commit{}
	return commits, nil
}