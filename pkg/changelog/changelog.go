// pkg/changelog/changelog.go
package changelog

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"buildy/pkg/config"
)

func GenerateChangelogs(cfg *config.Config, outputDir string) error {
	// Open the Git repository
	repo, err := git.PlainOpen(".")
	if err != nil {
		return fmt.Errorf("error opening Git repository: %v", err)
	}

	// Get the latest commit
	commit, err := getLatestCommit(repo)
	if err != nil {
		return fmt.Errorf("error getting latest commit: %v", err)
	}

	// Generate changelog for each subproject
	for _, subProject := range cfg.SubProjects {
		err := generateSubProjectChangelog(subProject, outputDir, commit)
		if err != nil {
			return fmt.Errorf("error generating changelog for subproject %s: %v", subProject.Name, err)
		}
	}

	// Generate centralized changelog
	err = generateCentralizedChangelog(cfg, outputDir, commit)
	if err != nil {
		return fmt.Errorf("error generating centralized changelog: %v", err)
	}

	return nil
}

func generateSubProjectChangelog(subProject config.SubProject, outputDir string, commit *object.Commit) error {
	changelogFile := filepath.Join(outputDir, subProject.Name, "CHANGELOG.md")

	// Create the directory if it doesn't exist
	err := os.MkdirAll(filepath.Dir(changelogFile), os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating directory for subproject changelog: %v", err)
	}

	// Check if the changelog file already exists
	_, err = os.Stat(changelogFile)
	if os.IsNotExist(err) {
		// If the file doesn't exist, create a new one with the initial content
		err := ioutil.WriteFile(changelogFile, []byte(fmt.Sprintf("# Changelog - %s\n\n", subProject.Name)), 0644)
		if err != nil {
			return fmt.Errorf("error creating subproject changelog file: %v", err)
		}
	}

	// Read the existing changelog content
	content, err := ioutil.ReadFile(changelogFile)
	if err != nil {
		return fmt.Errorf("error reading subproject changelog file: %v", err)
	}

	// Generate the new changelog entry
	entry := generateChangelogEntry(subProject, commit)

	// Prepend the new entry to the existing content
	updatedContent := fmt.Sprintf("%s\n%s", entry, string(content))

	// Write the updated content back to the changelog file
	err = ioutil.WriteFile(changelogFile, []byte(updatedContent), 0644)
	if err != nil {
		return fmt.Errorf("error writing subproject changelog file: %v", err)
	}

	return nil
}

func generateCentralizedChangelog(cfg *config.Config, outputDir string, commit *object.Commit) error {
	changelogFile := filepath.Join(outputDir, "CHANGELOG.md")

	// Check if the changelog file already exists
	_, err := os.Stat(changelogFile)
	if os.IsNotExist(err) {
		// If the file doesn't exist, create a new one with the initial content
		err := ioutil.WriteFile(changelogFile, []byte("# Centralized Changelog\n\n"), 0644)
		if err != nil {
			return fmt.Errorf("error creating centralized changelog file: %v", err)
		}
	}

	// Read the existing changelog content
	content, err := ioutil.ReadFile(changelogFile)
	if err != nil {
		return fmt.Errorf("error reading centralized changelog file: %v", err)
	}

	// Generate the new changelog entry
	entry := generateCentralizedChangelogEntry(cfg, commit)

	// Prepend the new entry to the existing content
	updatedContent := fmt.Sprintf("%s\n%s", entry, string(content))

	// Write the updated content back to the changelog file
	err = ioutil.WriteFile(changelogFile, []byte(updatedContent), 0644)
	if err != nil {
		return fmt.Errorf("error writing centralized changelog file: %v", err)
	}

	return nil
}

func generateChangelogEntry(subProject config.SubProject, commit *object.Commit) string {
	timestamp := commit.Author.When.Format("2006-01-02")
	entry := fmt.Sprintf("## [%s] - %s\n", subProject.Version, timestamp)
	entry += fmt.Sprintf("- %s\n\n", commit.Message)
	return entry
}

func generateCentralizedChangelogEntry(cfg *config.Config, commit *object.Commit) string {
	timestamp := commit.Author.When.Format("2006-01-02")
	entry := fmt.Sprintf("## [%s] - %s\n\n", cfg.Version, timestamp)

	for _, subProject := range cfg.SubProjects {
		entry += fmt.Sprintf("### %s\n", subProject.Name)
		entry += fmt.Sprintf("- Updated to version %s\n", subProject.Version)
	}

	entry += fmt.Sprintf("\nCommit: %s\n", commit.Hash.String())
	entry += fmt.Sprintf("Author: %s\n", commit.Author.Name)
	entry += fmt.Sprintf("Date: %s\n", timestamp)
	entry += fmt.Sprintf("Message: %s\n\n", commit.Message)

	return entry
}

func getLatestCommit(repo *git.Repository) (*object.Commit, error) {
	head, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("error getting HEAD reference: %v", err)
	}

	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return nil, fmt.Errorf("error getting latest commit: %v", err)
	}

	return commit, nil
}