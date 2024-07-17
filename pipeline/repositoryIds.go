package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
)

type RepositoryId int64

func ContainsRepositoryId(repositoryIds []RepositoryId, repositoryId RepositoryId) bool {
	_, found := slices.BinarySearch(repositoryIds, repositoryId)
	return found
}

func SaveRepositoryIds(repositoryIds []RepositoryId, outputFilePath string) error {
	bytes, err := json.Marshal(repositoryIds)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outputFilePath), os.ModePerm); err != nil {
		return err
	}
	if err := os.WriteFile(outputFilePath, bytes, 0644); err != nil {
		return err
	}

	return nil
}

func LoadRepositoryIds(inputFilePath string) ([]RepositoryId, error) {
	bytes, err := os.ReadFile(inputFilePath)
	if err != nil {
		return nil, err
	}

	var result []RepositoryId
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return nil, err
	}

	slices.Sort(result)
	return result, nil
}
