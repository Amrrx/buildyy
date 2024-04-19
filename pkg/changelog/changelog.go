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

	"github.com/go-git/go-git/plumbing/storer"
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
	lastSubProjectCommits, err := getLastSubProjectCommitsFromChangelog(centralizedChangelogFile, mainRepo)
	if err != nil {
		return fmt.Errorf("error getting last subproject commits from centralized changelog: %v", err)
	}

	// Create a slice to store the subproject commits in the correct order
	subProjectCommits := make([]*object.Commit, len(cfg.SubProjects))

	// Generate changelog for each subproject
	for i, subProject := range cfg.SubProjects {
		subProjectDir := filepath.Join(subProject.Path)

		// Check if the subproject has its own Git repository
		subProjectRepo, err := git.PlainOpen(subProjectDir)
		if err != nil {
			// If the subproject doesn't have its own repository, use the main repository
			subProjectRepo = mainRepo
		}

		// Get the last checked commit for the subproject from the centralized changelog
		lastSubProjectCommit := lastSubProjectCommits[subProject.Name]

		// Get the latest commit from the subproject repository
		latestSubProjectCommit, err := getLastCommit(subProjectRepo)
		if err != nil {
			return fmt.Errorf("error getting latest commit for subproject %s: %v", subProject.Name, err)
		}

		// If no commits are saved in the centralized changelog, get the first commit from the subproject repository
		if lastSubProjectCommit == nil {
			fmt.Println("No commits, first time")
			firstSubProjectCommit, err := getFirstCommit(subProjectRepo)
			if err != nil {
				return fmt.Errorf("error getting first commit for subproject %s: %v", subProject.Name, err)
			}
			lastSubProjectCommit = firstSubProjectCommit
		}

		// Store the latest subproject commit in the slice
		subProjectCommits[i] = latestSubProjectCommit

		// Generate the subproject changelog
		err = generateSubProjectChangelog(subProject, subProjectDir, subProjectRepo, lastSubProjectCommit, latestSubProjectCommit)
		if err != nil {
			return fmt.Errorf("error generating changelog for subproject %s: %v", subProject.Name, err)
		}
	}

	// Update the centralized changelog
	err = updateCentralizedChangelog(centralizedChangelogFile, cfg.SubProjects, mainRepo, cfg.Version, subProjectCommits)
	if err != nil {
		return fmt.Errorf("error updating centralized changelog: %v", err)
	}
	return nil
}

func updateCentralizedChangelog(changelogFile string, subProjects []config.SubProject, repo *git.Repository, centralVersion string, subProjectCommits []*object.Commit) error {
	// Read the existing changelog content
	content, _ := ioutil.ReadFile(changelogFile)

	// Get the current date
	currentDate := time.Now().Format("2006-01-02")

	// Generate the new changelog entry
	entry := fmt.Sprintf("## [%s] - %s\n\n", centralVersion, currentDate)
	for i, subProject := range subProjects {
		entry += fmt.Sprintf("### %s\n", subProject.Name)
		if subProjectCommits[i] != nil {
			entry += fmt.Sprintf("- Updated to version %s | %s\n", subProject.Version, subProjectCommits[i].Hash.String())
		} else {
			entry += fmt.Sprintf("- Updated to version %s\n", subProject.Version)
		}
	}

	// Get the latest commit from the main repository
	latestCentralCommit, err := getLastCommit(repo)
	if err != nil {
		return fmt.Errorf("error getting latest commit: %v", err)
	}

	// Get the commits up to the latest central commit
	centralCommits, err := getCommitsUpTo(repo, latestCentralCommit)
	if err != nil {
		return fmt.Errorf("error getting commits for central repository: %v", err)
	}

	// Add the central repository commits to the changelog entry
	if len(centralCommits) > 0 {
		entry += "\n### Central Repository\n"
		for _, commit := range centralCommits {
			entry += fmt.Sprintf("- %s\n", commit.Message)
		}
	}

	// Generate the commit information
	var commitInfo string
	if latestCentralCommit != nil {
		commitInfo = fmt.Sprintf("Commit: %s\n", latestCentralCommit.Hash.String())
		commitInfo += fmt.Sprintf("Author: %s\n", latestCentralCommit.Author.Name)
		commitInfo += fmt.Sprintf("Date: %s\n", latestCentralCommit.Author.When.Format("2006-01-02"))
		commitInfo += fmt.Sprintf("Message: %s\n", latestCentralCommit.Message)
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

	// Combine the entry with the existing content
	updatedContent := fmt.Sprintf("%s\n%s\n%s", entry, commitInfo, string(content))

	// Write the updated content back to the changelog file
	err = ioutil.WriteFile(changelogFile, []byte(updatedContent), 0644)
	if err != nil {
		return fmt.Errorf("error writing centralized changelog file: %v", err)
	}

	return nil
}

// func updateCentralizedChangelog(changelogFile string, subProjects []config.SubProject, latestCommit *object.Commit, centralVersion string) error {
// 	// Read the existing changelog content
// 	content, _ := ioutil.ReadFile(changelogFile)

// 	// Get the current date
// 	currentDate := time.Now().Format("2006-01-02")

// 	// Generate the new changelog entry
// 	entry := fmt.Sprintf("## [%s] - %s\n\n", centralVersion, currentDate)
// 	for _, subProject := range subProjects {
// 		entry += fmt.Sprintf("### %s\n", subProject.Name)
// 		entry += fmt.Sprintf("- Updated to version %s | %s\n", subProject.Version, latestCommit.Hash.String())
// 	}

// 	// Generate the commit information
// 	var commitInfo string
// 	if latestCommit != nil {
// 		commitInfo = fmt.Sprintf("Commit: %s\n", latestCommit.Hash.String())
// 		commitInfo += fmt.Sprintf("Author: %s\n", latestCommit.Author.Name)
// 		commitInfo += fmt.Sprintf("Date: %s\n", latestCommit.Author.When.Format("2006-01-02"))
// 		commitInfo += fmt.Sprintf("Message: %s\n", latestCommit.Message)
// 	} else {
// 		// Use default values if no commit data is found
// 		buildNumber := os.Getenv("BUILD_NUMBER")
// 		if buildNumber == "" {
// 			buildNumber = "Unknown"
// 		}
// 		hostname, _ := os.Hostname()
// 		commitInfo = fmt.Sprintf("Commit: %s\n", buildNumber)
// 		commitInfo += fmt.Sprintf("Author: %s\n", hostname)
// 		commitInfo += fmt.Sprintf("Date: %s\n", currentDate)
// 		commitInfo += "Message: Build Message <Recommend>\n"
// 	}

// 	// Combine the entry and commit information
// 	updatedContent := fmt.Sprintf("%s\n%s\n%s", entry, commitInfo, string(content))

// 	// Write the updated content back to the changelog file
// 	err := ioutil.WriteFile(changelogFile, []byte(updatedContent), 0644)
// 	if err != nil {
// 		return fmt.Errorf("error writing centralized changelog file: %v", err)
// 	}

// 	return nil
// }

func generateSubProjectChangelog(subProject config.SubProject, subProjectDir string, repo *git.Repository, lastSubProjectCommit, latestSubProjectCommit *object.Commit) error {
	changelogFile := filepath.Join(subProjectDir, "CHANGELOG.md")

	// Create the directory if it doesn't exist
	err := os.MkdirAll(filepath.Dir(changelogFile), os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating directory for subproject changelog: %v", err)
	}

	// Get the commits between the last subproject commit and the latest central commit
	commits, err := getCommitsBetween(repo, lastSubProjectCommit, latestSubProjectCommit)
	if err != nil {
		return fmt.Errorf("error getting commits for subproject %s: %v", subProject.Name, err)
	}

	// Generate the changelog entry for the subproject
	var entry string
	buildNumber := subProject.Version
	if len(commits) > 0 {
		latestCommit := commits[0]
		timestamp := latestCommit.Author.When.Format("2006-01-02")
		entry = fmt.Sprintf("## [%s] - %s\n", buildNumber, timestamp)
		for _, commit := range commits {
			entry += fmt.Sprintf("- %s\n", commit.Message)
		}
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

func getLastSubProjectCommitsFromChangelog(changelogFile string, repo *git.Repository) (map[string]*object.Commit, error) {
	lastSubProjectCommits := make(map[string]*object.Commit)

	// Read the centralized changelog file
	content, err := ioutil.ReadFile(changelogFile)
	if err != nil {
		if os.IsNotExist(err) {
			// If the file doesn't exist, return an empty map
			return lastSubProjectCommits, nil
		}
		return nil, fmt.Errorf("error reading centralized changelog file: %v", err)
	}

	// Parse the last subproject commit hashes from the changelog
	lines := strings.Split(string(content), "\n")
	subProject := ""
	for _, line := range lines {
		if strings.HasPrefix(line, "### ") {
			subProject = strings.TrimSpace(line[4:])
		} else if strings.Contains(line, " | ") {
			parts := strings.Split(line, " | ")
			if len(parts) == 2 {
				commitHash := strings.TrimSpace(parts[1])
				commit, err := repo.CommitObject(plumbing.NewHash(commitHash))
				if err == nil {
					lastSubProjectCommits[subProject] = commit
				}
			}
		}
	}

	return lastSubProjectCommits, nil
}

func getCommitsBetween(repo *git.Repository, oldCommit, newCommit *object.Commit) ([]*object.Commit, error) {
	var commits []*object.Commit

	// Create a new commit iterator
	iter, err := repo.Log(&git.LogOptions{From: newCommit.Hash})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	// Iterate through the commits until reaching the old commit
	err = iter.ForEach(func(commit *object.Commit) error {
		if commit.Hash == oldCommit.Hash {
			return storer.ErrStop
		}
		commits = append(commits, commit)
		return nil
	})
	if err != nil && err != storer.ErrStop {
		return nil, err
	}

	// Reverse the order of commits to get the oldest commit first
	for i, j := 0, len(commits)-1; i < j; i, j = i+1, j-1 {
		commits[i], commits[j] = commits[j], commits[i]
	}

	return commits, nil
}

func getFirstCommit(repo *git.Repository) (*object.Commit, error) {
	iter, err := repo.Log(&git.LogOptions{Order: git.LogOrderCommitterTime})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var firstCommit *object.Commit
	err = iter.ForEach(func(commit *object.Commit) error {
		firstCommit = commit
		return storer.ErrStop
	})
	if err != nil && err != storer.ErrStop {
		return nil, err
	}

	return firstCommit, nil
}

func getLastCommit(repo *git.Repository) (*object.Commit, error) {
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

func getCommitsUpTo(repo *git.Repository, latestCommit *object.Commit) ([]*object.Commit, error) {
	var commits []*object.Commit

	// Create a new commit iterator
	iter, err := repo.Log(&git.LogOptions{From: latestCommit.Hash})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	// Iterate through the commits up to the latest commit
	err = iter.ForEach(func(commit *object.Commit) error {
		commits = append(commits, commit)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Reverse the order of commits to get the oldest commit first
	for i, j := 0, len(commits)-1; i < j; i, j = i+1, j-1 {
		commits[i], commits[j] = commits[j], commits[i]
	}

	return commits, nil
}
