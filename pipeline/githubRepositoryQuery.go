package main

import (
	"fmt"
	"strings"
)

type GitHubRepositoryQuery struct {
	Language  *string
	CreatedAt *DateRange
	Stars     *Range
	Size      *Range
	Topic     *string
}

func (ghrso GitHubRepositoryQuery) ToString() string {
	var queryBuilder strings.Builder

	if ghrso.Language != nil {
		queryBuilder.WriteString(fmt.Sprintf("language:%s ", *ghrso.Language))
	}

	if ghrso.CreatedAt != nil {
		queryBuilder.WriteString(fmt.Sprintf("created:%s ", ghrso.CreatedAt.ToString()))
	}

	if ghrso.Stars != nil {
		queryBuilder.WriteString(fmt.Sprintf("stars:%s ", ghrso.Stars.ToString()))
	}

	if ghrso.Size != nil {
		queryBuilder.WriteString(fmt.Sprintf("size:%s ", ghrso.Size.ToString()))
	}

	if ghrso.Topic != nil {
		queryBuilder.WriteString(fmt.Sprintf("topic:%s ", *ghrso.Topic))
	}

	return strings.TrimSpace(queryBuilder.String())
}

func (ghrso GitHubRepositoryQuery) SplitByCreatedAt(numParts int) ([]GitHubRepositoryQuery, error) {
	createdAtRanges, err := ghrso.CreatedAt.Split(numParts)
	if err != nil {
		return nil, err
	}

	results := make([]GitHubRepositoryQuery, 0, numParts)

	for _, createdAtRange := range createdAtRanges {
		result := ghrso
		result.CreatedAt = &createdAtRange
		results = append(results, result)
	}

	return results, nil
}

func (ghrso GitHubRepositoryQuery) SplitByStars(numParts int) ([]GitHubRepositoryQuery, error) {
	starsRanges, err := ghrso.Stars.Split(numParts)
	if err != nil {
		return nil, err
	}

	results := make([]GitHubRepositoryQuery, 0, numParts)

	for _, starsRange := range starsRanges {
		result := ghrso
		result.Stars = &starsRange
		results = append(results, result)
	}

	return results, nil
}

func (ghrso GitHubRepositoryQuery) SplitBySize(numParts int) ([]GitHubRepositoryQuery, error) {
	sizeRanges, err := ghrso.Size.Split(numParts)
	if err != nil {
		return nil, err
	}

	results := make([]GitHubRepositoryQuery, 0, numParts)

	for _, sizeRange := range sizeRanges {
		result := ghrso
		result.Size = &sizeRange
		results = append(results, result)
	}

	return results, nil
}

func (ghrso GitHubRepositoryQuery) Split(numParts int) ([]GitHubRepositoryQuery, error) {
	if numParts <= 0 {
		return nil, fmt.Errorf("cannot split GitHubRepositoryQuery into %d parts", numParts)
	}

	if numParts == 1 {
		return []GitHubRepositoryQuery{ghrso}, nil
	}

	results := make([]GitHubRepositoryQuery, 0, numParts)
	results = append(results, ghrso)
	for len(results) < numParts {
		firstResult := results[0]

		if firstResult.CreatedAt != nil {
			numDays := firstResult.CreatedAt.NumDays()
			numRemainingParts := (numParts - len(results)) + 1
			if numDays >= numRemainingParts {
				result, err := firstResult.SplitByCreatedAt(numRemainingParts)
				if err != nil {
					return nil, err
				}
				results = append(results[1:], result...)
				continue
			} else if numDays > 1 {
				result, err := firstResult.SplitByCreatedAt(numDays)
				if err != nil {
					return nil, err
				}
				results = append(results[1:], result...)
				continue
			}
		}

		if firstResult.Size != nil && firstResult.Size.NumItems() >= 2 {
			result, err := firstResult.SplitBySize(2)
			if err != nil {
				return nil, err
			}
			results = append(results[1:], result...)
			continue
		}

		if firstResult.Stars != nil && firstResult.Stars.NumItems() >= 2 {
			result, err := firstResult.SplitByStars(2)
			if err != nil {
				return nil, err
			}
			results = append(results[1:], result...)
			continue
		}

		return nil, fmt.Errorf("cannot split query \"%s\" into %d parts", ghrso.ToString(), numParts)
	}

	return results, nil
}
