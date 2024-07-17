package main

func UniqueSliceElements[T comparable](inputSlice []T) []T {
	uniqueSlice := make([]T, 0, len(inputSlice))
	seen := make(map[T]bool, len(inputSlice))
	for _, element := range inputSlice {
		if !seen[element] {
			uniqueSlice = append(uniqueSlice, element)
			seen[element] = true
		}
	}
	return uniqueSlice
}

func IntersectionSlice[T comparable](lSlice, rSlice []T) []T {
	intersection := make([]T, 0, len(lSlice))
	for _, lElement := range lSlice {
		for _, rElement := range rSlice {
			if lElement == rElement {
				intersection = append(intersection, lElement)
				break
			}
		}
	}
	return UniqueSliceElements(intersection)
}
