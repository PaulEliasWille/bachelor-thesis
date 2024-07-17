package main

import "fmt"

type Range struct {
	Start        int
	ExclusiveEnd int
}

func (r Range) NumItems() int {
	return r.ExclusiveEnd - r.Start
}

func (r Range) Split(numParts int) ([]Range, error) {
	if numParts <= 0 {
		return nil, fmt.Errorf("cannot split Range into %d parts", numParts)
	}

	if numParts == 1 {
		return []Range{r}, nil
	}

	numItems := r.NumItems()
	if numParts > numItems {
		return nil, fmt.Errorf("cannot split Range into more parts than items")
	}

	partSize := numItems / numParts

	ranges := make([]Range, 0, numItems)
	for i := 0; i < numParts; i++ {
		startOffset := i * partSize
		exclusiveEndOffset := (i + 1) * partSize
		if i == numParts-1 {
			exclusiveEndOffset = numItems
		}

		start := r.Start + startOffset
		exclusiveEnd := r.Start + exclusiveEndOffset

		ranges = append(ranges, Range{
			Start:        start,
			ExclusiveEnd: exclusiveEnd,
		})
	}

	return ranges, nil
}

func (r Range) ToString() string {
	return fmt.Sprintf("%d..%d", r.Start, r.ExclusiveEnd-1)
}
