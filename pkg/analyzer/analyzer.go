package analyzer

import (
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"buildy/pkg/config"
)

func DetermineVersionIncrement(subProject config.SubProject, repo *git.Repository, lastCheckpointCommit, latestCommit *object.Commit) (string, error) {
	// Get the commits between the last checkpoint and the latest commit for the subproject
	commits, err := getCommitsForSubproject(repo, subProject, lastCheckpointCommit, latestCommit)
	if err != nil {
		return "", err
	}

	// Analyze the commits to determine the version increment
	increment := analyzeCommits(commits)

	return increment, nil
}

func getCommitsForSubproject(repo *git.Repository, subProject config.SubProject, lastCheckpointCommit, latestCommit *object.Commit) ([]*object.Commit, error) {
	var commits []*object.Commit

	// If there is no last checkpoint commit, retrieve all commits for the subproject up to the latest commit
	if lastCheckpointCommit == nil {
		commitIter, err := repo.Log(&git.LogOptions{From: latestCommit.Hash, PathFilter: func(path string) bool {
			return strings.HasPrefix(path, subProject.Path)
		}})
		if err != nil {
			return nil, err
		}

		err = commitIter.ForEach(func(commit *object.Commit) error {
			commits = append(commits, commit)
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		// Retrieve commits for the subproject between the last checkpoint commit and the latest commit
		commitIter, err := repo.Log(&git.LogOptions{From: latestCommit.Hash, Since: &lastCheckpointCommit.Author.When, PathFilter: func(path string) bool {
			return strings.HasPrefix(path, subProject.Path)
		}})
		if err != nil {
			return nil, err
		}

		err = commitIter.ForEach(func(commit *object.Commit) error {
			if commit.Hash != lastCheckpointCommit.Hash {
				commits = append(commits, commit)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return commits, nil
}

func analyzeCommits(commits []*object.Commit) string {
	// Analyze the commits to determine the version increment
	var increment string

	for _, commit := range commits {
		message := commit.Message
		if strings.Contains(message, "BREAKING CHANGE") || strings.Contains(message, "!:") {
			increment = "major"
			break
		}
		if strings.Contains(message, "feat:") || strings.Contains(message, "feature:") {
			increment = "minor"
		}
	}

	// If no specific increment is determined, default to "patch"
	if increment == "" {
		increment = "patch"
	}

	return increment
}