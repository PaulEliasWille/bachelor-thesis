package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/github"
)

func LoadRepositoryQueries(inputFilePath string) ([]string, error) {
	bytes, err := os.ReadFile(inputFilePath)
	if err != nil {
		return nil, err
	}

	var result []string
	if json.Unmarshal(bytes, &result) != nil {
		return nil, err
	}

	return result, nil
}

func ChunkGitHubRepositoryQuery(initialQuery GitHubRepositoryQuery, numWorkers int, outputFile string) {
	workQueue := make(chan GitHubRepositoryQuery, 1000)
	resultQueue := make(chan GitHubRepositoryQuery, 1000)

	var wgWorker sync.WaitGroup
	var wgReceiver sync.WaitGroup
	var wgWorkQueue sync.WaitGroup

	for workerIndex := 0; workerIndex < numWorkers; workerIndex++ {
		wgWorker.Add(1)
		go func() {
			githubClient := MakeGitHubClient("proxies.txt", workerIndex)
			defer wgWorker.Done()
			for query := range workQueue {
				numRepositories := githubClient.GetNumRepositories(query.ToString())
				if numRepositories <= 1000 {
					resultQueue <- query
				} else {
					numParts := numRepositories / 1000
					if numRepositories%1000 != 0 {
						numParts++
					}
					results, err := query.Split(numParts)
					if err != nil {
						fmt.Printf("Warn: failed to split query \"%s\": %v\n", query.ToString(), err)
					} else {
						for _, result := range results {
							wgWorkQueue.Add(1)
							workQueue <- result
						}
					}
				}
				wgWorkQueue.Done()
			}
		}()
	}

	wgReceiver.Add(1)
	go func() {
		defer wgReceiver.Done()

		queries := make([]GitHubRepositoryQuery, 0)
		for query := range resultQueue {
			queries = append(queries, query)
		}

		slices.SortFunc(queries, func(lhs, rhs GitHubRepositoryQuery) int {
			switch {
			case lhs.CreatedAt == nil && rhs.CreatedAt == nil:
				return 0
			case lhs.CreatedAt == nil:
				return -1
			case rhs.CreatedAt == nil:
				return 1
			case lhs.CreatedAt.Start.IsBefore(rhs.CreatedAt.Start):
				return -1
			case rhs.CreatedAt.Start.IsBefore(lhs.CreatedAt.Start):
				return 1
			default:
				return 0
			}
		})

		queryStrings := make([]string, 0)
		for _, query := range queries {
			queryStrings = append(queryStrings, query.ToString())
		}

		queriesBytes, _ := json.Marshal(queryStrings)
		if err := os.WriteFile(outputFile, queriesBytes, 0644); err != nil {
			panic(err)
		}
	}()

	wgWorkQueue.Add(1)
	workQueue <- initialQuery

	wgWorkQueue.Wait()

	close(workQueue)
	close(resultQueue)

	wgWorker.Wait()
	wgReceiver.Wait()
}

func ScrapeGitHub(queries []string, numWorkers int, outputDirectory string) {
	workQueue := make(chan string, 10)
	resultQueue := make(chan []github.Repository, 10)

	var wgWorker sync.WaitGroup
	var wgReceiver sync.WaitGroup
	var wgWorkQueue sync.WaitGroup

	for workerIndex := 0; workerIndex < numWorkers; workerIndex++ {
		wgWorker.Add(1)
		go func() {
			githubClient := MakeGitHubClient("proxies.txt", workerIndex)
			defer wgWorker.Done()
			for query := range workQueue {
				resultQueue <- githubClient.GetRepositories(query)
				wgWorkQueue.Done()
			}
		}()
	}

	wgReceiver.Add(1)
	go func() {
		defer wgReceiver.Done()

		if err := os.MkdirAll(outputDirectory, os.ModePerm); err != nil {
			panic(err)
		}
		for repositories := range resultQueue {
			for _, repository := range repositories {
				repositoryId := repository.GetID()

				repositoryOutputFile := filepath.Join(
					outputDirectory,
					fmt.Sprintf("%d.json", repositoryId),
				)

				repositoriesBytes, err := json.Marshal(repository)
				if err != nil {
					panic(err)
				}

				if err := os.WriteFile(repositoryOutputFile, repositoriesBytes, 0644); err != nil {
					panic(err)
				}
			}
		}
	}()

	for _, query := range queries {
		wgWorkQueue.Add(1)
		workQueue <- query
	}
	wgWorkQueue.Wait()

	close(workQueue)
	close(resultQueue)

	wgWorker.Wait()
	wgReceiver.Wait()
}

func Unzip(zippedBytes []byte, dest string) error {
	zippedBytesReader := bytes.NewReader(zippedBytes)
	r, err := zip.NewReader(zippedBytesReader, int64(len(zippedBytes)))
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		nameParts := strings.SplitN(f.Name, "/", 2)
		if len(nameParts) != 2 {
			return fmt.Errorf("invalid file name format: %s", f.Name)
		}

		outPath := filepath.Join(dest, nameParts[1])

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(outPath, f.Mode()); err != nil {
				return err
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(outPath), f.Mode()); err != nil {
				return err
			}
			f, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func DownloadRepositoryFile(httpClient *http.Client, fileURL, outPath string) error {
	response, err := httpClient.Get(fileURL)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(response.Body)

	if response.StatusCode == http.StatusTooManyRequests {
		if retryAfter, err := ParseRetryAfterHeader(response.Header.Get("Retry-After")); err == nil {
			return &ThrottledError{retryAfter: retryAfter}
		}
		return &ThrottledError{retryAfter: 60 * time.Second}
	} else if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outPath), os.ModePerm); err != nil {
		return err
	}

	if err := os.WriteFile(outPath, responseBytes, 0644); err != nil {
		return err
	}

	return nil
}

func DownloadRepositoryFiles(
	numWorkers int, repositoryIds []RepositoryId,
	repositoryInfosDirectory string, filePaths []string,
	outputDirectory string,
	resume bool,
) {
	workQueue := make(chan RepositoryId, 10_000)

	var wgWorker sync.WaitGroup
	var wgWorkQueue sync.WaitGroup

	for workerIndex := 0; workerIndex < numWorkers; workerIndex++ {
		wgWorker.Add(1)
		go func() {
			defer wgWorker.Done()

			httpClient := NewHTTPClient("proxies.txt", workerIndex)
			for repositoryId := range workQueue {
				repositoryInfoPath := path.Join(
					repositoryInfosDirectory,
					fmt.Sprintf("%d.json", repositoryId),
				)
				repositoryInfo, err := LoadRepositoryInfo(repositoryInfoPath)
				if err != nil {
					fmt.Printf("Error loading repository %d info: %s\n", repositoryId, err)
					continue
				}

				repositoryOutputDirectory := path.Join(
					outputDirectory,
					fmt.Sprintf("%d", repositoryId),
				)

				if err := os.MkdirAll(repositoryOutputDirectory, os.ModePerm); err != nil {
					fmt.Printf("Error creating output directory %s: %s\n", repositoryOutputDirectory, err)
					continue
				}

				for _, filePath := range filePaths {
					repositoryFileOutputPath := path.Join(
						repositoryOutputDirectory,
						fmt.Sprintf("%s", filePath),
					)

					fileURL := fmt.Sprintf(
						"https://raw.githubusercontent.com/%s/%s/%s",
						repositoryInfo.GetFullName(),
						repositoryInfo.GetDefaultBranch(),
						filePath,
					)

					if err := RetryWithoutResult(func() error {
						return DownloadRepositoryFile(httpClient, fileURL, repositoryFileOutputPath)
					}, DefaultErrorHandler); err != nil {
						// fmt.Printf("Error downloading %s for repository %d: %s\n", fileURL, repositoryId, err)
						continue
					}
				}

				wgWorkQueue.Done()
			}
		}()
	}

	ignoredRepositoryIds := make(map[RepositoryId]struct{})
	if resume {
		files, err := os.ReadDir(outputDirectory)
		if err == nil {
			for _, file := range files {
				if !file.IsDir() {
					continue
				}

				if repositoryId, err := strconv.ParseInt(file.Name(), 10, 64); err == nil {
					ignoredRepositoryIds[RepositoryId(repositoryId)] = struct{}{}
				}
			}
		}
	}

	fmt.Printf("Ignoring %d repositories\n", len(ignoredRepositoryIds))

	for _, repositoryId := range repositoryIds {
		if _, ok := ignoredRepositoryIds[repositoryId]; ok {
			continue
		}
		wgWorkQueue.Add(1)
		workQueue <- repositoryId
	}
	fmt.Printf("wait wgWorkQueue\n")
	wgWorkQueue.Wait()
	fmt.Printf("waited wgWorkQueue\n")
	fmt.Printf("close workQueue\n")
	close(workQueue)
	fmt.Printf("closed workQueue\n")
	fmt.Printf("wait wgWorker\n")
	wgWorker.Wait()
	fmt.Printf("waited wgWorker\n")
}

func DownloadRepository(archiveURL string, path string) error {
	response, err := http.Get(archiveURL)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(response.Body)

	responseBytes, err := io.ReadAll(response.Body)

	if err := Unzip(responseBytes, path); err != nil {
		return err
	}

	return nil
}

func DownloadRepositories(repositoryIds []RepositoryId, repositoryInfosDirectory string, outputDirectory string) error {
	for _, repositoryId := range repositoryIds {
		repositoryInfoPath := path.Join(repositoryInfosDirectory, fmt.Sprintf("%d.json", repositoryId))
		repositoryInfo, err := LoadRepositoryInfo(repositoryInfoPath)
		if err != nil {
			fmt.Printf("Error loading repository %d info: %s\n", repositoryId, err)
			continue
		}

		repositoryPath := path.Join(outputDirectory, fmt.Sprintf("%d", repositoryId))

		if err := os.MkdirAll(repositoryPath, os.ModePerm); err != nil {
			fmt.Printf("Error loading repository %d info: %s\n", repositoryId, err)
			continue
		}

		archiveURL := fmt.Sprintf("%s/archive/refs/heads/%s.zip", repositoryInfo.GetHTMLURL(), repositoryInfo.GetDefaultBranch())
		if err := DownloadRepository(archiveURL, repositoryPath); err != nil {
			fmt.Printf("Error cloning repository %d: %s\n", repositoryId, err)
			continue
		}
	}

	return nil
}
