package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"
)

func RepositoriesDataToCSV(repositoriesData []RepositoryData, outPath string) error {
	var buffer bytes.Buffer
	csvWriter := csv.NewWriter(&buffer)

	_ = csvWriter.Write([]string{
		"id",
		"url",
		"name",
		"description",
		"num_stars",
		"num_forks",
		"num_issues",
		"num_commits",
		"num_contributors",
		"active_days",
		"last_commit_at",
		"num_functions",
		"used_platforms",
		"used_frameworks",
		"num_packages",
		"num_published_packages",
		"num_faas_handlers",
		"num_faas_runtime_dependencies",
	})

	for _, repositoryData := range repositoriesData {
		_ = csvWriter.Write([]string{
			fmt.Sprintf("%d", repositoryData.RepositoryId),
			repositoryData.Url,
			repositoryData.Name,
			repositoryData.Description,
			fmt.Sprintf("%d", repositoryData.Stars),
			fmt.Sprintf("%d", repositoryData.Forks),
			fmt.Sprintf("%d", repositoryData.NumIssues),
			fmt.Sprintf("%d", repositoryData.NumCommits),
			fmt.Sprintf("%d", repositoryData.NumContributors),
			fmt.Sprintf("%d", repositoryData.ActiveHumanDays),
			repositoryData.LastHumanCommitAt.Format(time.RFC3339),
			fmt.Sprintf("%d", repositoryData.NumFunctions),
			usedPlatformsToString(repositoryData.UsedPlatforms),
			usedFrameworksToString(repositoryData.UsedFrameworks),
			fmt.Sprintf("%d", repositoryData.NumPackages),
			fmt.Sprintf("%d", repositoryData.NumPublishedToNPM),
			fmt.Sprintf("%d", repositoryData.NumFaaSHandlers),
			fmt.Sprintf("%d", repositoryData.NumFaaSRuntimeDependencies),
		})
	}

	csvWriter.Flush()

	if err := csvWriter.Error(); err != nil {
		return err
	}

	if err := os.MkdirAll(path.Dir(outPath), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(outPath, buffer.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}

func RepositoriesDataToJSON(repositoriesData []RepositoryData, outPath string) error {
	repositoriesDataBytes, err := json.Marshal(repositoriesData)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outPath), os.ModePerm); err != nil {
		return err
	}
	if err := os.WriteFile(outPath, repositoriesDataBytes, 0644); err != nil {
		return err
	}

	return nil
}

type RepositoryStatistics struct {
	TotalNumApplications int

	TotalNumFunctions                 int
	TotalNumFunctionsByPlatform       map[FaaSPlatform]int
	TotalNumFunctionsByFramework      map[FaaSFramework]int
	TotalNumFunctionsByLocation       map[FaaSLocation]int
	TotalNumFunctionsByInvocationType map[FaaSInvocationType]int

	TotalNumFunctionsByFrameworkByPlatform      map[FaaSPlatform]map[FaaSFramework]int
	TotalNumFunctionsByLocationByPlatform       map[FaaSPlatform]map[FaaSLocation]int
	TotalNumFunctionsByInvocationTypeByPlatform map[FaaSPlatform]map[FaaSInvocationType]int

	TotalNumFunctionsByPlatformByFramework       map[FaaSFramework]map[FaaSPlatform]int
	TotalNumFunctionsByLocationByFramework       map[FaaSFramework]map[FaaSLocation]int
	TotalNumFunctionsByInvocationTypeByFramework map[FaaSFramework]map[FaaSInvocationType]int

	TotalNumFunctionsByFrameworkByInvocationType map[FaaSInvocationType]map[FaaSFramework]int

	AverageNumFunctionsPerApplication float64
	MinNumFunctionsPerApplication     int
	MaxNumFunctionsPerApplication     int
}

func RepositoriesDataStatisticsToJSON(repositoriesData []RepositoryData, outPath string) error {
	result := RepositoryStatistics{
		MinNumFunctionsPerApplication:                1000,
		MaxNumFunctionsPerApplication:                -1000,
		TotalNumFunctionsByPlatform:                  make(map[FaaSPlatform]int),
		TotalNumFunctionsByFramework:                 make(map[FaaSFramework]int),
		TotalNumFunctionsByLocation:                  make(map[FaaSLocation]int),
		TotalNumFunctionsByInvocationType:            make(map[FaaSInvocationType]int),
		TotalNumFunctionsByFrameworkByPlatform:       make(map[FaaSPlatform]map[FaaSFramework]int),
		TotalNumFunctionsByLocationByPlatform:        make(map[FaaSPlatform]map[FaaSLocation]int),
		TotalNumFunctionsByInvocationTypeByPlatform:  make(map[FaaSPlatform]map[FaaSInvocationType]int),
		TotalNumFunctionsByPlatformByFramework:       make(map[FaaSFramework]map[FaaSPlatform]int),
		TotalNumFunctionsByLocationByFramework:       make(map[FaaSFramework]map[FaaSLocation]int),
		TotalNumFunctionsByInvocationTypeByFramework: make(map[FaaSFramework]map[FaaSInvocationType]int),
		TotalNumFunctionsByFrameworkByInvocationType: make(map[FaaSInvocationType]map[FaaSFramework]int),
	}
	for _, data := range repositoriesData {
		result.TotalNumApplications += 1

		result.TotalNumFunctions += data.NumFunctions
		result.AverageNumFunctionsPerApplication += float64(data.NumFunctions)
		result.MinNumFunctionsPerApplication = min(result.MinNumFunctionsPerApplication, data.NumFunctions)
		result.MaxNumFunctionsPerApplication = max(result.MaxNumFunctionsPerApplication, data.NumFunctions)

		for _, function := range data.Functions {
			result.TotalNumFunctionsByPlatform[function.Platform] += 1
			result.TotalNumFunctionsByFramework[function.Framework] += 1
			result.TotalNumFunctionsByLocation[function.Location] += 1
			result.TotalNumFunctionsByInvocationType[function.InvocationType] += 1

			if _, ok := result.TotalNumFunctionsByFrameworkByPlatform[function.Platform]; !ok {
				result.TotalNumFunctionsByFrameworkByPlatform[function.Platform] = make(map[FaaSFramework]int)
			}

			if _, ok := result.TotalNumFunctionsByLocationByPlatform[function.Platform]; !ok {
				result.TotalNumFunctionsByLocationByPlatform[function.Platform] = make(map[FaaSLocation]int)
			}

			if _, ok := result.TotalNumFunctionsByInvocationTypeByPlatform[function.Platform]; !ok {
				result.TotalNumFunctionsByInvocationTypeByPlatform[function.Platform] = make(map[FaaSInvocationType]int)
			}

			result.TotalNumFunctionsByFrameworkByPlatform[function.Platform][function.Framework] += 1
			result.TotalNumFunctionsByLocationByPlatform[function.Platform][function.Location] += 1
			result.TotalNumFunctionsByInvocationTypeByPlatform[function.Platform][function.InvocationType] += 1

			if _, ok := result.TotalNumFunctionsByPlatformByFramework[function.Framework]; !ok {
				result.TotalNumFunctionsByPlatformByFramework[function.Framework] = make(map[FaaSPlatform]int)
			}

			if _, ok := result.TotalNumFunctionsByLocationByFramework[function.Framework]; !ok {
				result.TotalNumFunctionsByLocationByFramework[function.Framework] = make(map[FaaSLocation]int)
			}

			if _, ok := result.TotalNumFunctionsByInvocationTypeByFramework[function.Framework]; !ok {
				result.TotalNumFunctionsByInvocationTypeByFramework[function.Framework] = make(map[FaaSInvocationType]int)
			}

			result.TotalNumFunctionsByPlatformByFramework[function.Framework][function.Platform] += 1
			result.TotalNumFunctionsByLocationByFramework[function.Framework][function.Location] += 1
			result.TotalNumFunctionsByInvocationTypeByFramework[function.Framework][function.InvocationType] += 1

			if _, ok := result.TotalNumFunctionsByFrameworkByInvocationType[function.InvocationType]; !ok {
				result.TotalNumFunctionsByFrameworkByInvocationType[function.InvocationType] = make(map[FaaSFramework]int)
			}

			result.TotalNumFunctionsByFrameworkByInvocationType[function.InvocationType][function.Framework] += 1
		}
	}
	result.AverageNumFunctionsPerApplication = result.AverageNumFunctionsPerApplication / float64(result.TotalNumApplications)

	resultBytes, err := json.Marshal(result)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outPath), os.ModePerm); err != nil {
		return err
	}
	if err := os.WriteFile(outPath, resultBytes, 0644); err != nil {
		return err
	}

	return nil

}

func ExportRepositories(
	repositoryIds []RepositoryId,
	repositoriesDataDirectory string,
	outDirectory string,
) error {
	repositories := make([]RepositoryData, 0)
	faasRepositories := make([]RepositoryData, 0)
	nonFaasRepositories := make([]RepositoryData, 0)

	for _, repositoryId := range repositoryIds {
		repositoryData, err := LoadRepositoryData(path.Join(
			repositoriesDataDirectory,
			fmt.Sprintf("%d.json", repositoryId),
		))
		if err != nil {
			fmt.Printf("Failed to load repository %d: %s\n", repositoryId, err)
			continue
		}

		repositories = append(repositories, repositoryData)

		if repositoryData.NumFunctions > 0 {
			faasRepositories = append(faasRepositories, repositoryData)
		} else {
			nonFaasRepositories = append(nonFaasRepositories, repositoryData)
		}
	}

	if err := RepositoriesDataToCSV(repositories, path.Join(outDirectory, "repositories.csv")); err != nil {
		return err
	}

	if err := RepositoriesDataToCSV(faasRepositories, path.Join(outDirectory, "faasRepositories.csv")); err != nil {
		return err
	}

	if err := RepositoriesDataToCSV(nonFaasRepositories, path.Join(outDirectory, "nonFaasRepositories.csv")); err != nil {
		return err
	}

	if err := RepositoriesDataToJSON(repositories, path.Join(outDirectory, "repositories.json")); err != nil {
		return err
	}

	if err := RepositoriesDataToJSON(faasRepositories, path.Join(outDirectory, "faasRepositories.json")); err != nil {
		return err
	}

	if err := RepositoriesDataToJSON(nonFaasRepositories, path.Join(outDirectory, "nonFaasRepositories.json")); err != nil {
		return err
	}

	if err := RepositoriesDataStatisticsToJSON(repositories, path.Join(outDirectory, "repositoriesStatistics.json")); err != nil {
		return err
	}

	if err := RepositoriesDataStatisticsToJSON(faasRepositories, path.Join(outDirectory, "faasRepositoriesStatistics.json")); err != nil {
		return err
	}

	if err := RepositoriesDataStatisticsToJSON(nonFaasRepositories, path.Join(outDirectory, "nonFaasRepositoriesStatistics.json")); err != nil {
		return err
	}

	return nil
}

//func ExportRepositoryPackages(
//	repositoryIds []RepositoryId,
//	repositoriesDataDirectory string,
//	filter func(RepositoryData, RepositoryPackageData) bool,
//	outPath string,
//) error {
//	var buffer bytes.Buffer
//	csvWriter := csv.NewWriter(&buffer)
//
//	_ = csvWriter.Write([]string{
//		"repo_id",
//		"repo_url",
//		"repo_name",
//		"repo_description",
//		"repo_stars",
//		"repo_active_days",
//		"app_path",
//		"app_name",
//		"app_description",
//	})
//
//	for _, repositoryId := range repositoryIds {
//		repositoryData, err := LoadRepositoryData(path.Join(
//			repositoriesDataDirectory,
//			fmt.Sprintf("%d.json", repositoryId),
//		))
//		if err != nil {
//			fmt.Printf("Failed to load repository %d: %s\n", repositoryId, err)
//			continue
//		}
//
//		for _, packageData := range repositoryData.Packages {
//			if !filter(repositoryData, packageData) {
//				continue
//			}
//
//			_ = csvWriter.Write([]string{
//				fmt.Sprintf("%d", repositoryData.RepositoryId),
//				repositoryData.Url,
//				repositoryData.Name,
//				repositoryData.Description,
//				fmt.Sprintf("%d", repositoryData.Stars),
//				fmt.Sprintf("%d", repositoryData.ActiveHumanDays),
//				packageData.RootPath,
//				packageData.Name,
//				packageData.Description,
//			})
//		}
//	}
//
//	csvWriter.Flush()
//
//	if err := csvWriter.Error(); err != nil {
//		return err
//	}
//
//	if err := os.MkdirAll(path.Dir(outPath), 0755); err != nil {
//		return err
//	}
//
//	if err := os.WriteFile(outPath, buffer.Bytes(), 0644); err != nil {
//		return err
//	}
//
//	return nil
//}
