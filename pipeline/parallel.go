package main

import (
	"net/http"
	"sync"
)

func ProcessInParallel[T any, R any](
	items []T,
	itemProcessor func(item T, httpClient *http.Client, githubClient GitHubClient) (result R, ok bool),
	resultProcessor func(result R),
	numItemProcessors int,
	numResultProcessors int,
	itemBufferSize int,
	resultBufferSize int,
) {
	itemQueue := make(chan T, itemBufferSize)
	resultQueue := make(chan R, resultBufferSize)

	var wgItemProcessors sync.WaitGroup
	var wgResultProcessors sync.WaitGroup
	var wgItemQueue sync.WaitGroup

	for itemProcessorIndex := 0; itemProcessorIndex < numItemProcessors; itemProcessorIndex++ {
		wgItemProcessors.Add(1)
		go func() {
			httpClient := NewHTTPClient("proxies.txt", itemProcessorIndex)
			githubClient := MakeGitHubClient("proxies.txt", itemProcessorIndex)
			defer wgItemProcessors.Done()
			for item := range itemQueue {
				result, ok := itemProcessor(item, httpClient, githubClient)
				if ok {
					resultQueue <- result
				}
				wgItemQueue.Done()
			}
		}()
	}

	for resultProcessorIndex := 0; resultProcessorIndex < numResultProcessors; resultProcessorIndex++ {
		wgResultProcessors.Add(1)
		go func() {
			defer wgResultProcessors.Done()
			for result := range resultQueue {
				resultProcessor(result)
			}
		}()
	}

	for _, item := range items {
		wgItemQueue.Add(1)
		itemQueue <- item
	}
	wgItemQueue.Wait()

	close(itemQueue)
	close(resultQueue)

	wgItemProcessors.Wait()
	wgResultProcessors.Wait()
}
