package main

import (
	"context"
	"errors"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

func loadFile(inPath string) (string, error) {
	bytes, err := os.ReadFile(inPath)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(bytes)), nil
}

func aggregateFiles(directory string, paths []string) string {
	var sb strings.Builder
	sb.WriteString("# CONTEXT\n\n")
	for idx, path := range paths {
		content, err := loadFile(filepath.Join(directory, path))
		if err != nil {
			continue
		}
		sb.WriteString(content)
		if idx != len(paths)-1 {
			sb.WriteString("\n---\n")
		}
	}

	return sb.String()
}

func SummarizeRepository(
	client *openai.Client,
	repositoryId RepositoryId,
	repositoriesDirectory string,
	documentationFiles []string,
) (string, error) {
	repositoryDirectory := filepath.Join(repositoriesDirectory, fmt.Sprintf("%d", repositoryId))

	documentationFilesContent := aggregateFiles(repositoryDirectory, documentationFiles)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo0125,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleSystem,
					Content: "# INSTRUCTIONS\n\n" +
						"You will be given a list of project files. Your task is to SUMMARIZE WHAT THE PROJECT DOES in 50 words or less.\n" +
						"The summary SHOULD:\n" +
						"- describe what the project does\n" +
						"- answer if the project uses a serverless architecture (e.g. aws lambda)\n" +
						"The summary SHOULD NOT:\n" +
						"- contain licensing info\n" +
						"- contain attributions or setup/configuration instructions\n" +
						"- other, irrelevant, details\n" +
						"Keep your answer concise and only answer with the summary and nothing else.\n" +
						"Do not make up information, if you do not know the answer just say so.\n" +
						"If you are unable to create a summary, just say: \"not enough information\".",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: documentationFilesContent,
				},
			},
		},
	)

	var openaiErr *openai.APIError
	switch {
	case errors.Is(err, nil):
		return resp.Choices[0].Message.Content, nil
	case errors.As(err, &openaiErr):
		if openaiErr.HTTPStatusCode == http.StatusTooManyRequests {
			time.Sleep(30 * time.Second)
			return SummarizeRepository(client, repositoryId, repositoryDirectory, documentationFiles)
		}
		fallthrough
	default:
		return "", err
	}
}

func SummarizeRepositories(
	repositoryIds []RepositoryId,
	repositoriesDirectory string,
	documentationFiles []string,
	outDirectory string,
	resume bool,
) error {
	client, err := NewOpenAIClient()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(outDirectory, os.ModePerm); err != nil {
		return err
	}

	ignoredRepositoryIds := make(map[RepositoryId]struct{})
	if resume {
		files, err := os.ReadDir(outDirectory)
		if err == nil {
			for _, file := range files {
				if file.IsDir() {
					continue
				}

				if repositoryId, err := strconv.ParseInt(strings.TrimSuffix(file.Name(), ".txt"), 10, 64); err == nil {
					ignoredRepositoryIds[RepositoryId(repositoryId)] = struct{}{}
				}
			}
		}
	}

	fmt.Printf("Ignoring %d repositories\n", len(ignoredRepositoryIds))

	for idx, repositoryId := range repositoryIds {
		if _, ok := ignoredRepositoryIds[repositoryId]; ok {
			continue
		}

		fmt.Printf("Processing repository %d\n", idx)

		summary, err := SummarizeRepository(client, repositoryId, repositoriesDirectory, documentationFiles)
		if err != nil {
			fmt.Printf("Error summarizing repository %d: %s\n", repositoryId, err)
			continue
		}

		outPath := filepath.Join(outDirectory, fmt.Sprintf("%d.txt", repositoryId))
		if err := os.WriteFile(outPath, []byte(summary), os.ModePerm); err != nil {
			fmt.Printf("Failed to write summary to %s: %s\n", outPath, err)
			continue
		}
	}
	return nil
}

func buildCategorizePromptContext(
	repositoryId RepositoryId,
	repositoryInfosDirectory string,
	repositorySummariesDirectory string,
) string {
	repositoryInfoPath := filepath.Join(repositoryInfosDirectory, fmt.Sprintf("%d.json", repositoryId))
	repositorySummaryPath := filepath.Join(repositorySummariesDirectory, fmt.Sprintf("%d.txt", repositoryId))

	var sb strings.Builder
	sb.WriteString("# CONTEXT\n\n")

	if info, err := LoadRepositoryInfo(repositoryInfoPath); err == nil {
		sb.WriteString(fmt.Sprintf("Name: %s\n", info.GetName()))
		sb.WriteString(fmt.Sprintf("Full Name: %s\n", info.GetFullName()))
		sb.WriteString(fmt.Sprintf("Description: %s\n", info.GetDescription()))
	}

	if summary, err := loadFile(repositorySummaryPath); err == nil {
		sb.WriteString(fmt.Sprintf("Summary of readme: %s\n", summary))
	}

	return sb.String()
}

func pickMajority(choices []openai.ChatCompletionChoice) (string, int, error) {
	if len(choices) == 0 {
		return "", 0, fmt.Errorf("no choices found")
	}

	choiceCounts := map[string]int{}
	for _, choice := range choices {
		choiceCounts[choice.Message.Content] += 1
	}

	bestCount := -1
	bestChoice := ""
	for choice, count := range choiceCounts {
		if count > bestCount {
			bestChoice = choice
			bestCount = count
		}
	}

	numBestChoices := 0
	for _, count := range choiceCounts {
		if count == bestCount {
			numBestChoices += 1
		}
	}

	if numBestChoices != 1 {
		return "", 0, fmt.Errorf("no unique majority found")
	}

	return bestChoice, bestCount, nil
}

func CategorizeRepository(
	client *openai.Client,
	repositoryId RepositoryId,
	repositoryInfosDirectory string,
	repositorySummariesDirectory string,
	retry int,
	maxRetries int,
) (string, error) {
	if retry > maxRetries {
		return "", fmt.Errorf("maximum number of retries exceeded")
	}

	prompt := "# INSTRUCTIONS\n\n" +
		"You will be details about a project. Your task is to CATEGORIZE THE PROJECT based on this information. " +
		"Possible categories are:\n" +
		"- \"real_application\": which means the project is an application that could be used " +
		"in the real world. Real world means that it is NOT a demo or example.\n" +
		"- \"demo_application\": which means that the project is an application that serves " +
		"as a demo or example and is NOT intended as a real world application.\n" +
		"- \"other\": which means that the project is no application but, for example, a framework or library.\n" +
		"Always try to pick the most descriptive category. Do not make up new categories. If you are unable to decide use \"other\".\n" +
		"Only answer with the category and nothing else.\n"
	promptContext := buildCategorizePromptContext(repositoryId, repositoryInfosDirectory, repositorySummariesDirectory)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       openai.GPT3Dot5Turbo,
			N:           5,
			Temperature: 0.0,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: prompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: promptContext,
				},
			},
		},
	)

	var openaiErr *openai.APIError
	switch {
	case errors.Is(err, nil):
		category, _, err := pickMajority(resp.Choices)
		if err != nil {
			fmt.Printf("Error picking majority: %v\n", err)
			return CategorizeRepository(client, repositoryId, repositoryInfosDirectory, repositorySummariesDirectory, retry+1, maxRetries)
		}

		if !slices.Contains([]string{"real_application", "demo_application", "other"}, category) {
			fmt.Printf("Invalid category: %s\n", category)
			return CategorizeRepository(client, repositoryId, repositoryInfosDirectory, repositorySummariesDirectory, retry+1, maxRetries)
		}
		return resp.Choices[0].Message.Content, nil
	case errors.As(err, &openaiErr):
		if openaiErr.HTTPStatusCode == http.StatusTooManyRequests {
			fmt.Printf("Throttled, retrying in 60 seconds...\n")
			time.Sleep(60 * time.Second)
			return CategorizeRepository(client, repositoryId, repositoryInfosDirectory, repositorySummariesDirectory, retry+1, maxRetries)
		}
		fallthrough
	default:
		return "", err
	}
}

func CategorizeRepositories(
	repositoryIds []RepositoryId,
	repositoryInfosDirectory string,
	repositorySummariesDirectory string,
	outDirectory string,
	resume bool,
) error {
	client, err := NewOpenAIClient()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(outDirectory, os.ModePerm); err != nil {
		return err
	}

	ignoredRepositoryIds := make(map[RepositoryId]struct{})
	if resume {
		files, err := os.ReadDir(outDirectory)
		if err == nil {
			for _, file := range files {
				if file.IsDir() {
					continue
				}

				if repositoryId, err := strconv.ParseInt(strings.TrimSuffix(file.Name(), ".txt"), 10, 64); err == nil {
					ignoredRepositoryIds[RepositoryId(repositoryId)] = struct{}{}
				}
			}
		}
	}

	fmt.Printf("Ignoring %d repositories\n", len(ignoredRepositoryIds))

	for idx, repositoryId := range repositoryIds {
		if _, ok := ignoredRepositoryIds[repositoryId]; ok {
			continue
		}

		fmt.Printf("Processing repository %d\n", idx)

		category, err := CategorizeRepository(client, repositoryId, repositoryInfosDirectory, repositorySummariesDirectory, 0, 1)
		if err != nil {
			fmt.Printf("Error categorizing repository %d: %s\n", repositoryId, err)
			continue
		}

		outPath := filepath.Join(outDirectory, fmt.Sprintf("%d.txt", repositoryId))
		if err := os.WriteFile(outPath, []byte(category), os.ModePerm); err != nil {
			fmt.Printf("Failed to write category to %s: %s\n", outPath, err)
			continue
		}
	}
	return nil
}
