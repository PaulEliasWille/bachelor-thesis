package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"

	"github.com/google/go-github/github"
)

func LoadRepositoryInfo(inputPath string) (github.Repository, error) {
	repositoryBytes, err := os.ReadFile(inputPath)
	if err != nil {
		return github.Repository{}, err
	}

	repositoryInfo := github.Repository{}
	err = json.Unmarshal(repositoryBytes, &repositoryInfo)
	if err != nil {
		return github.Repository{}, err
	}

	return repositoryInfo, nil
}

func AggregateRepositoryIds(inputDirectory string, outputFilePath string) error {
	matches, err := filepath.Glob(filepath.Join(inputDirectory, "*.json"))
	if err != nil {
		return err
	}

	resultSet := map[RepositoryId]struct{}{}
	for _, match := range matches {
		repositoriesBytes, err := os.ReadFile(match)
		if err != nil {
			return err
		}

		var repository github.Repository
		err = json.Unmarshal(repositoriesBytes, &repository)
		if err != nil {
			return err
		}

		resultSet[RepositoryId(repository.GetID())] = struct{}{}
	}

	result := make([]RepositoryId, 0, len(resultSet))
	for repositoryId := range resultSet {
		result = append(result, repositoryId)
	}

	slices.Sort(result)

	return SaveRepositoryIds(result, outputFilePath)
}
