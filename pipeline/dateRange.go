package main

import "fmt"

type DateRange struct {
	Start        Date
	ExclusiveEnd Date
}

func (dr DateRange) NumDays() int {
	return dr.Start.DaysUntil(dr.ExclusiveEnd)
}

func (dr DateRange) Split(numParts int) ([]DateRange, error) {
	if numParts <= 0 {
		return nil, fmt.Errorf("cannot split DateRange into %d parts", numParts)
	}

	if numParts == 1 {
		return []DateRange{dr}, nil
	}

	numDays := dr.NumDays()

	if numParts > numDays {
		return nil, fmt.Errorf("cannot split DateRange into more parts than days")
	}

	partSize := numDays / numParts

	ranges := make([]DateRange, 0, numParts)
	for i := 0; i < numParts; i++ {
		startOffset := i * partSize
		exclusiveEndOffset := (i + 1) * partSize
		if i == numParts-1 {
			exclusiveEndOffset = numDays
		}

		start := dr.Start.AddDays(startOffset)
		exclusiveEnd := dr.Start.AddDays(exclusiveEndOffset)

		ranges = append(ranges, DateRange{
			Start:        start,
			ExclusiveEnd: exclusiveEnd,
		})
	}

	return ranges, nil
}

func (dr DateRange) ToString() string {
	return fmt.Sprintf("%s..%s", dr.Start.ToString(), dr.ExclusiveEnd.Previous().ToString())
}

func (dr DateRange) ToList() []Date {
	dates := make([]Date, 0)

	date := dr.Start
	for date.IsBefore(dr.ExclusiveEnd) {
		dates = append(dates, date)
		date = date.Next()
	}
	return dates
}
