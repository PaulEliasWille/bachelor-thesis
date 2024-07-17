package main

import (
	"fmt"
	"os"
	"path"
	"slices"
	"time"
)

func FilterRelevantRepositoryIds(
	repositoryIds []RepositoryId,
	keywords []string,
	excludeKeywords []string,
	documentationFiles []string,
	repositoryInfoPath string,
	repositoryEventsPath string,
	repositoriesPath string,
) []RepositoryId {
	results := make([]RepositoryId, 0)
	for _, repositoryId := range repositoryIds {
		info, err := LoadRepositoryInfo(
			path.Join(repositoryInfoPath, fmt.Sprintf("%d.json", repositoryId)),
		)
		if err != nil {
			fmt.Printf("Error loading repository info: %v\n", err)
			continue
		}

		if info.GetArchived() {
			continue
		}

		if info.GetPushedAt().Before(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)) {
			continue
		}

		ageEstimation := info.GetPushedAt().Sub(info.GetCreatedAt().Time)
		minAge := time.Hour * 8766 // 24 hours * 365.25 days = 8766 hours per year
		if ageEstimation < minAge {
			continue
		}

		if len(info.GetDescription()) == 0 {
			continue
		}

		if checkForKeywords(info.GetFullName(), excludeKeywords) > 0 {
			continue
		}

		if checkForKeywords(info.GetDescription(), excludeKeywords) > 0 {
			continue
		}

		foundExcludeKeyword := false
		for _, topic := range info.Topics {
			if checkForKeywords(topic, excludeKeywords) > 0 {
				foundExcludeKeyword = true
				break
			}
		}

		if foundExcludeKeyword {
			continue
		}

		numInfoMatches := checkForKeywords(info.GetFullName(), keywords)
		numInfoMatches += checkForKeywords(info.GetDescription(), keywords)
		for _, topic := range info.Topics {
			numInfoMatches += checkForKeywords(topic, keywords)
		}

		events, err := LoadRepositoryEventsOrDefault(
			path.Join(repositoryEventsPath, fmt.Sprintf("%d.json", repositoryId)),
			make([]RepositoryEvent, 0),
		)
		if err != nil {
			fmt.Printf("Error loading events: %v\n", err)
			continue
		}

		numEventMatches := 0
		for _, event := range events {
			numEventMatches += event.CountKeywordMatches(keywords)
		}

		numDocumentationMatches := 0
		for _, documentationFile := range documentationFiles {
			fileBytes, err := os.ReadFile(path.Join(
				repositoriesPath,
				fmt.Sprintf("%d", repositoryId),
				documentationFile,
			))
			if err != nil {
				continue
			}

			fileString := string(fileBytes)

			numDocumentationMatches += checkForKeywords(fileString, keywords)
		}

		if numInfoMatches >= 1 || numEventMatches+numDocumentationMatches >= 2 {
			results = append(results, repositoryId)
		}
	}
	return results
}

func FilterHighlyRelevantRepositoryIds(
	repositoryIds []RepositoryId,
	ignoreRepositoryIds []RepositoryId,
	repositoriesDataPath string,
) []RepositoryId {
	results := make([]RepositoryId, 0)
	for _, repositoryId := range repositoryIds {
		if slices.Contains(ignoreRepositoryIds, repositoryId) {
			continue
		}
		
		repositoryDataPath := path.Join(repositoriesDataPath, fmt.Sprintf("%d.json", repositoryId))
		repositoryData, err := LoadRepositoryData(repositoryDataPath)
		if err != nil {
			fmt.Printf("Error loading repository data: %v\n", err)
			continue
		}

		if repositoryData.NumIssues < 5 {
			continue
		}

		if repositoryData.NumCommits < 5 {
			continue
		}

		if repositoryData.ActiveHumanDays < 365 {
			continue
		}

		if repositoryData.LastHumanCommitAt.Before(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)) {
			continue
		}

		results = append(results, repositoryId)

	}
	return results
}
