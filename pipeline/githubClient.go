package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/github"
)

func extractTotalFromResponse[T any](responseItems []T, numRequestedItems int, response *github.Response) int {
	if len(responseItems) < numRequestedItems {
		return 0
	}

	link := response.Header.Get("Link")
	if len(link) == 0 {
		// NOTE: the GitHub API does not include the "Link" header if total < numRequestedItems.
		//       based on that we know that we already received all items -> total == len(responseItems)
		return len(responseItems)
	}

	segments := strings.Split(link, ",")
	for _, segment := range segments {
		subSegments := strings.Split(segment, ";")
		if len(subSegments) != 2 {
			continue
		}

		segmentRel := subSegments[1]
		segmentRel = strings.TrimSpace(segmentRel)
		segmentRel = strings.TrimPrefix(segmentRel, "rel=\"")
		segmentRel = strings.TrimSuffix(segmentRel, "\"")

		if segmentRel != "last" {
			continue
		}

		segmentUrl := subSegments[0]
		segmentUrl = strings.TrimSpace(segmentUrl)
		segmentUrl = strings.Trim(segmentUrl, "<>")

		parsedSegmentUrl, err := url.Parse(segmentUrl)
		if err != nil {
			continue
		}

		perPage, err := strconv.ParseInt(parsedSegmentUrl.Query().Get("per_page"), 10, 32)
		if err != nil {
			continue
		}

		page, err := strconv.ParseInt(parsedSegmentUrl.Query().Get("page"), 10, 32)
		if err != nil {
			continue
		}

		return int(page * perPage)
	}

	return -1
}

type GitHubClient struct {
	client *github.Client
}

func MakeGitHubClient(proxyFilePath string, proxyIndex int) GitHubClient {
	file, err := os.Open(proxyFilePath)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	proxyUrls := make([]string, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lineParts := strings.FieldsFunc(line, func(r rune) bool {
			return r == ':'
		})
		if len(lineParts) != 4 {
			panic(err)
		}
		proxyIp := lineParts[0]
		proxyPort := lineParts[1]
		proxyUser := lineParts[2]
		proxyPassword := lineParts[3]

		proxyUrls = append(proxyUrls, fmt.Sprintf("http://%s:%s@%s:%s", proxyUser, proxyPassword, proxyIp, proxyPort))
	}

	safeProxyIndex := proxyIndex % len(proxyUrls)
	proxyUrl, err := url.Parse(proxyUrls[safeProxyIndex])
	if err != nil {
		panic(err)
	}

	httpClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}

	return GitHubClient{client: github.NewClient(httpClient)}
}

func (ghc GitHubClient) GetRepositoriesPage(query string, page, perPage int) *github.RepositoriesSearchResult {
	for {
		opts := &github.SearchOptions{Sort: "stars", Order: "asc", ListOptions: github.ListOptions{PerPage: perPage, Page: page}}
		searchResult, _, err := ghc.client.Search.Repositories(context.Background(), query, opts)
		if err == nil {
			return searchResult
		}

		var githubRateLimitErr *github.RateLimitError

		switch {
		case errors.As(err, &githubRateLimitErr):
			backoff := time.Until(githubRateLimitErr.Rate.Reset.Time) + 10*time.Second // Add a buffer to avoid running into the rate limit again
			fmt.Printf("Rate limit exceeded. Retrying in %v seconds\n", backoff.Seconds())
			time.Sleep(backoff)
			fmt.Println("Retrying...")
		default:
			backoff := 10 * time.Second
			fmt.Printf("Error: %v. Retrying in %v seconds\n", err, backoff.Seconds())
			time.Sleep(backoff)
			fmt.Println("Retrying...")
		}
	}
}

func (ghc GitHubClient) GetNumRepositories(query string) int {
	return ghc.GetRepositoriesPage(query, 1, 1).GetTotal()
}

func (ghc GitHubClient) GetRepositories(query string) []github.Repository {
	const PageSize = 100

	repositories := make([]github.Repository, 0)

	searchResult := ghc.GetRepositoriesPage(query, 1, PageSize)
	repositories = append(repositories, searchResult.Repositories...)

	numRepositories := searchResult.GetTotal()
	if numRepositories > 1000 {
		fmt.Printf("Warn: query \"%s\" yielded %d results. The last %d result(s) will be skipped \n", query, numRepositories, numRepositories-1000)
		numRepositories = 1000
	}

	numPages := numRepositories / PageSize
	if numRepositories%PageSize != 0 {
		numPages++
	}

	for page := 2; page <= numPages; page++ {
		searchResult := ghc.GetRepositoriesPage(query, page, PageSize)
		repositories = append(repositories, searchResult.Repositories...)
	}

	return repositories
}

func (ghc GitHubClient) GetContributorsPage(owner, repo string, page, perPage int) ([]*github.Contributor, int) {
	for {
		opts := &github.ListContributorsOptions{
			ListOptions: github.ListOptions{Page: page, PerPage: perPage},
		}
		contributors, response, err := ghc.client.Repositories.ListContributors(
			context.Background(),
			owner,
			repo,
			opts,
		)
		if err == nil {
			return contributors, extractTotalFromResponse(contributors, perPage, response)
		}

		var githubRateLimitErr *github.RateLimitError

		switch {
		case errors.As(err, &githubRateLimitErr):
			backoff := time.Until(githubRateLimitErr.Rate.Reset.Time) + 10*time.Second // Add a buffer to avoid running into the rate limit again
			fmt.Printf("Rate limit exceeded. Retrying in %v seconds\n", backoff.Seconds())
			time.Sleep(backoff)
			fmt.Println("Retrying...")
		default:
			backoff := 10 * time.Second
			fmt.Printf("Error: %v. Retrying in %v seconds\n", err, backoff.Seconds())
			time.Sleep(backoff)
			fmt.Println("Retrying...")
		}
	}
}

func (ghc GitHubClient) GetContributors(owner, repo string) []*github.Contributor {
	const PageSize = 100

	contributors := make([]*github.Contributor, 0)
	page := 1
	for {
		newContributors, _ := ghc.GetContributorsPage(owner, repo, page, PageSize)
		contributors = append(contributors, newContributors...)

		if len(newContributors) != PageSize {
			break
		}
		page += 1
	}

	return contributors
}

func (ghc GitHubClient) GetCommitsPage(owner, repo string, page, perPage int) ([]*github.RepositoryCommit, int) {
	for {
		opts := &github.CommitsListOptions{
			ListOptions: github.ListOptions{Page: page, PerPage: perPage},
		}
		commits, response, err := ghc.client.Repositories.ListCommits(
			context.Background(),
			owner,
			repo,
			opts,
		)
		if err == nil {
			return commits, extractTotalFromResponse(commits, perPage, response)
		}

		var githubRateLimitErr *github.RateLimitError

		switch {
		case errors.As(err, &githubRateLimitErr):
			backoff := time.Until(githubRateLimitErr.Rate.Reset.Time) + 10*time.Second // Add a buffer to avoid running into the rate limit again
			fmt.Printf("Rate limit exceeded. Retrying in %v seconds\n", backoff.Seconds())
			time.Sleep(backoff)
			fmt.Println("Retrying...")
		default:
			backoff := 10 * time.Second
			fmt.Printf("Error: %v. Retrying in %v seconds\n", err, backoff.Seconds())
			time.Sleep(backoff)
			fmt.Println("Retrying...")
		}
	}
}

func (ghc GitHubClient) GetCommits(owner, repo string) []*github.RepositoryCommit {
	const PageSize = 100

	commits := make([]*github.RepositoryCommit, 0)
	page := 1
	for {
		newCommits, _ := ghc.GetCommitsPage(owner, repo, page, PageSize)
		commits = append(commits, newCommits...)

		if len(newCommits) != PageSize {
			break
		}
		page += 1
	}

	return commits
}

func (ghc GitHubClient) GetIssuesPage(owner, repo, state string, page, perPage int) ([]*github.Issue, int) {
	for {
		opts := &github.IssueListByRepoOptions{
			State:       state,
			ListOptions: github.ListOptions{Page: page, PerPage: perPage},
		}
		issues, response, err := ghc.client.Issues.ListByRepo(
			context.Background(),
			owner,
			repo,
			opts,
		)
		if err == nil {
			return issues, extractTotalFromResponse(issues, perPage, response)
		}

		var githubRateLimitErr *github.RateLimitError

		switch {
		case errors.As(err, &githubRateLimitErr):
			backoff := time.Until(githubRateLimitErr.Rate.Reset.Time) + 10*time.Second // Add a buffer to avoid running into the rate limit again
			fmt.Printf("Rate limit exceeded. Retrying in %v seconds\n", backoff.Seconds())
			time.Sleep(backoff)
			fmt.Println("Retrying...")
		default:
			backoff := 10 * time.Second
			fmt.Printf("Error: %v. Retrying in %v seconds\n", err, backoff.Seconds())
			time.Sleep(backoff)
			fmt.Println("Retrying...")
		}
	}
}

func (ghc GitHubClient) GetIssues(owner, repo, state string) []*github.Issue {
	const PageSize = 100

	issues := make([]*github.Issue, 0)
	page := 1
	for {
		newIssues, _ := ghc.GetIssuesPage(owner, repo, state, page, PageSize)
		issues = append(issues, newIssues...)

		if len(newIssues) != PageSize {
			break
		}
		page += 1
	}

	return issues
}
