package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/go-github/github"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func arrayOfPointersToValues[T any](items []*T) []T {
	values := make([]T, len(items))
	for i, v := range items {
		values[i] = *v
	}
	return values
}

func getNumIssues(githubClient GitHubClient, repositoryInfo *github.Repository) (numOpen int, numClosed int) {
	name := repositoryInfo.GetName()
	owner := strings.TrimSuffix(repositoryInfo.GetFullName(), fmt.Sprintf("/%s", name))

	_, numOpen = githubClient.GetIssuesPage(owner, name, "open", 1, 1)
	_, numClosed = githubClient.GetIssuesPage(owner, name, "closed", 1, 1)

	return numOpen, numClosed
}

func downloadRepositoryCommitsHeadAndTail(githubClient GitHubClient, repositoryInfo *github.Repository) ([]*github.RepositoryCommit, int) {
	const MinTailCommits = 15

	name := repositoryInfo.GetName()
	owner := strings.TrimSuffix(repositoryInfo.GetFullName(), fmt.Sprintf("/%s", name))

	commitsHead, estNumCommits := githubClient.GetCommitsPage(owner, name, 1, 100)
	estNumPages := estNumCommits / 100
	if estNumPages <= 1 || len(commitsHead) < 100 {
		return commitsHead, len(commitsHead)
	}

	commitsTail, _ := githubClient.GetCommitsPage(owner, name, estNumPages, 100)
	estNumCommitsError := 100 - len(commitsTail)
	numCommits := estNumCommits - estNumCommitsError

	if len(commitsTail) >= MinTailCommits || numCommits < 100+MinTailCommits {
		return append(commitsHead, commitsTail...), numCommits
	}

	commitsPreTail, _ := githubClient.GetCommitsPage(owner, name, estNumPages-1, 100)

	return append(append(commitsHead, commitsPreTail...), commitsTail...), numCommits
}

func getNumContributors(githubClient GitHubClient, repositoryInfo *github.Repository) int {
	name := repositoryInfo.GetName()
	owner := strings.TrimSuffix(repositoryInfo.GetFullName(), fmt.Sprintf("/%s", name))
	_, numContributors := githubClient.GetContributorsPage(owner, name, 1, 1)
	return numContributors
}

type RepositoryIssuesCommitsAndContributors struct {
	RepositoryId       RepositoryId
	NumOpenIssues      int
	NumClosedIssues    int
	CommitsHeadAndTail []github.RepositoryCommit
	NumCommits         int
	NumContributors    int
}

func SaveRepositoryIssuesCommitsAndContributors(ricc *RepositoryIssuesCommitsAndContributors, outPath string) error {
	bytes, err := json.Marshal(ricc)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outPath), os.ModePerm); err != nil {
		return err
	}
	if err := os.WriteFile(outPath, bytes, 0644); err != nil {
		return err
	}

	return nil
}

func LoadRepositoryIssuesCommitsAndContributors(inPath string) (RepositoryIssuesCommitsAndContributors, error) {
	resultBytes, err := os.ReadFile(inPath)
	if err != nil {
		return RepositoryIssuesCommitsAndContributors{}, err
	}

	var result RepositoryIssuesCommitsAndContributors
	err = json.Unmarshal(resultBytes, &result)
	if err != nil {
		return RepositoryIssuesCommitsAndContributors{}, err
	}

	return result, nil
}

func DownloadRepositoriesIssuesCommitsAndContributors(
	numWorkers int,
	repositoryIds []RepositoryId,
	repositoryInfosDirectory string,
	repositoryIssuesCommitsAndContributorsDirectory string,
	resume bool,
) {
	ProcessInParallel(
		repositoryIds,
		func(repositoryId RepositoryId, _ *http.Client, githubClient GitHubClient) (RepositoryIssuesCommitsAndContributors, bool) {
			if resume {
				if _, err := os.Stat(path.Join(
					repositoryIssuesCommitsAndContributorsDirectory,
					fmt.Sprintf("%d.json", repositoryId),
				)); err == nil {
					return RepositoryIssuesCommitsAndContributors{}, false
				}
			}

			repositoryInfo, err := LoadRepositoryInfo(path.Join(
				repositoryInfosDirectory,
				fmt.Sprintf("%d.json", repositoryId),
			))
			if err != nil {
				fmt.Printf("failed to load repository info for repository id %d: %v\n", repositoryId, err)
				return RepositoryIssuesCommitsAndContributors{}, false
			}

			numOpenIssues, numClosedIssues := getNumIssues(githubClient, &repositoryInfo)
			commitsHeadAndTail, numCommits := downloadRepositoryCommitsHeadAndTail(githubClient, &repositoryInfo)
			numContributors := getNumContributors(githubClient, &repositoryInfo)

			return RepositoryIssuesCommitsAndContributors{
				RepositoryId:       repositoryId,
				NumOpenIssues:      numOpenIssues,
				NumClosedIssues:    numClosedIssues,
				CommitsHeadAndTail: arrayOfPointersToValues(commitsHeadAndTail),
				NumCommits:         numCommits,
				NumContributors:    numContributors,
			}, true
		},
		func(data RepositoryIssuesCommitsAndContributors) {
			fmt.Printf("processed repository %d\n", data.RepositoryId)
			if err := SaveRepositoryIssuesCommitsAndContributors(
				&data,
				path.Join(repositoryIssuesCommitsAndContributorsDirectory, fmt.Sprintf("%d.json", data.RepositoryId)),
			); err != nil {
				fmt.Printf("error saving repository issues: %v\n", err)
			}
		},
		numWorkers,
		1,
		10_000,
		10_000,
	)
}
