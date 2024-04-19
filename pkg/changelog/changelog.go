// pkg/changelog/changelog.go
package changelog

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"buildy/pkg/config"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)


func GenerateChangelogs(cfg *config.Config, outputDir string) error {
	// Open the main Git repository
	mainRepo, err := git.PlainOpen(".")
	if err != nil {
		return fmt.Errorf("error opening main Git repository: %v", err)
	}

	// Read the centralized changelog file
	centralizedChangelogFile := filepath.Join(outputDir, "CHANGELOG.md")
	_, err = getLastCheckedCommitFromChangelog(centralizedChangelogFile)
	if err != nil {
		return fmt.Errorf("error getting last checked commit from centralized changelog: %v", err)
	}

	// Get the latest commit from the main repository
	latestCommit, err := getLatestCommit(mainRepo)
	if err != nil {
		return fmt.Errorf("error getting latest commit: %v", err)
	}

	// Generate changelog for each subproject
	for _, subProject := range cfg.SubProjects {
		subProjectDir := filepath.Join(subProject.Path)

		// Check if the subproject has its own Git repository
		
		subProjectRepo, err := git.PlainOpen(subProjectDir)
		if err == nil {
			// If the subproject has its own repository, use it to generate the changelog
			err = generateSubProjectChangelog(subProject, subProjectDir, subProjectRepo, latestCommit)
			if err != nil {
				return fmt.Errorf("error generating changelog for subproject %s: %v", subProject.Name, err)
			}
		} else {
			// If the subproject doesn't have its own repository, use the main repository
			err = generateSubProjectChangelog(subProject, subProjectDir, mainRepo, latestCommit)
			if err != nil {
				return fmt.Errorf("error generating changelog for subproject %s: %v", subProject.Name, err)
			}
		}
	}

	// Update the centralized changelog with the latest commit
	err = updateCentralizedChangelog(centralizedChangelogFile, cfg.SubProjects, latestCommit, cfg.Version)
	if err != nil {
		return fmt.Errorf("error updating centralized changelog: %v", err)
	}

	return nil
}

func generateSubProjectChangelog(subProject config.SubProject, subProjectDir string, repo *git.Repository, latestCommit *object.Commit) error {
	changelogFile := filepath.Join(subProjectDir, "CHANGELOG.md")

	// Create the directory if it doesn't exist
	err := os.MkdirAll(filepath.Dir(changelogFile), os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating directory for subproject changelog: %v", err)
	}

	// Get the commits since the last checked commit
	commits, err := getCommitsSinceCheckpoint(repo, nil)
	if err != nil {
		return fmt.Errorf("error getting commits for subproject %s: %v", subProject.Name, err)
	}

	// Generate the changelog entry for the subproject
	var entry string
	buildNumber := subProject.Version
	timestamp := latestCommit.Author.When.Format("2006-01-02")
	entry = fmt.Sprintf("## [%s] - %s\n", buildNumber, timestamp)
	for _, commit := range commits {
		entry += fmt.Sprintf("- %s\n", commit.Message)
	}

	// Read the existing changelog content
	content, _ := ioutil.ReadFile(changelogFile)

	// Prepend the new entry to the existing content
	updatedContent := fmt.Sprintf("%s\n%s", entry, string(content))

	// Write the updated content back to the changelog file
	err = ioutil.WriteFile(changelogFile, []byte(updatedContent), 0644)
	if err != nil {
		return fmt.Errorf("error writing subproject changelog file: %v", err)
	}

	return nil
}

func generateChangelogEntry(subProject config.SubProject, commit *object.Commit) string {
	timestamp := commit.Author.When.Format("2006-01-02")
	entry := fmt.Sprintf("## [%s] - %s\n", subProject.Version, timestamp)
	entry += fmt.Sprintf("- %s\n", commit.Message)
	return entry
}

func getLastCheckedCommitFromChangelog(changelogFile string) (*object.Commit, error) {
	// Read the centralized changelog file
	content, err := ioutil.ReadFile(changelogFile)
	if err != nil {
		if os.IsNotExist(err) {
			// If the file doesn't exist, return nil as the last checked commit
			return nil, nil
		}
		return nil, fmt.Errorf("error reading centralized changelog file: %v", err)
	}

	// Parse the last checked commit hash from the changelog
	lines := strings.Split(string(content), "\n")
	if len(lines) > 0 {
		lastLine := lines[len(lines)-1]
		parts := strings.Split(lastLine, "|")
		if len(parts) > 1 {
			commitHash := strings.TrimSpace(parts[1])

			// Open the Git repository
			repo, err := git.PlainOpen(".")
			if err != nil {
				return nil, fmt.Errorf("error opening Git repository: %v", err)
			}

			// Get the commit object based on the commit hash
			commitObject, err := repo.CommitObject(plumbing.NewHash(commitHash))
			if err != nil {
				return nil, fmt.Errorf("error getting commit object: %v", err)
			}

			return commitObject, nil
		}
	}

	return nil, nil
}

func getCommitsSinceCheckpoint(repo *git.Repository, lastCheckedCommit *object.Commit) ([]*object.Commit, error) {
	var commits []*object.Commit

	// If there is no last checked commit, retrieve all commits
	if lastCheckedCommit == nil {
		commitIter, err := repo.Log(&git.LogOptions{})
		if err != nil {
			return nil, fmt.Errorf("error creating commit iterator: %v", err)
		}

		err = commitIter.ForEach(func(commit *object.Commit) error {
			commits = append(commits, commit)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error iterating commits: %v", err)
		}
	} else {
		// Retrieve commits since the last checked commit
		commitIter, err := repo.Log(&git.LogOptions{Since: &lastCheckedCommit.Author.When})
		if err != nil {
			return nil, fmt.Errorf("error creating commit iterator: %v", err)
		}

		err = commitIter.ForEach(func(commit *object.Commit) error {
			if commit.Hash != lastCheckedCommit.Hash {
				commits = append(commits, commit)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error iterating commits: %v", err)
		}
	}

	return commits, nil
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

func updateCentralizedChangelog(changelogFile string, subProjects []config.SubProject, latestCommit *object.Commit, centralVersion string) error {
	// Read the existing changelog content
	content, _ := ioutil.ReadFile(changelogFile)

	// Get the current date
	currentDate := time.Now().Format("2006-01-02")

	// Generate the new changelog entry
	entry := fmt.Sprintf("## [%s] - %s\n\n", centralVersion, currentDate)
	for _, subProject := range subProjects {
		entry += fmt.Sprintf("### %s\n", subProject.Name)
		entry += fmt.Sprintf("- Updated to version %s | %s\n", subProject.Version, latestCommit.Hash.String())
	}

	// Generate the commit information
	var commitInfo string
	if latestCommit != nil {
		commitInfo = fmt.Sprintf("Commit: %s\n", latestCommit.Hash.String())
		commitInfo += fmt.Sprintf("Author: %s\n", latestCommit.Author.Name)
		commitInfo += fmt.Sprintf("Date: %s\n", latestCommit.Author.When.Format("2006-01-02"))
		commitInfo += fmt.Sprintf("Message: %s\n", latestCommit.Message)
	} else {
		// Use default values if no commit data is found
		buildNumber := os.Getenv("BUILD_NUMBER")
		if buildNumber == "" {
			buildNumber = "Unknown"
		}
		hostname, _ := os.Hostname()
		commitInfo = fmt.Sprintf("Commit: %s\n", buildNumber)
		commitInfo += fmt.Sprintf("Author: %s\n", hostname)
		commitInfo += fmt.Sprintf("Date: %s\n", currentDate)
		commitInfo += "Message: Build Message <Recommend>\n"
	}

	// Combine the entry and commit information
	updatedContent := fmt.Sprintf("%s\n%s\n%s", entry, commitInfo, string(content))

	// Write the updated content back to the changelog file
	err := ioutil.WriteFile(changelogFile, []byte(updatedContent), 0644)
	if err != nil {
		return fmt.Errorf("error writing centralized changelog file: %v", err)
	}

	return nil
}
