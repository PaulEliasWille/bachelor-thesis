package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type RepositoryEventId int64

type RepositoryEvent struct {
	RepositoryId RepositoryId
	EventId      RepositoryEventId
	EventType    string
	Event        map[string]interface{}
}

func ParseRepositoryEvent(rawEvent []byte) (RepositoryEvent, error) {
	var event map[string]interface{}
	if err := json.Unmarshal(rawEvent, &event); err != nil {
		return RepositoryEvent{}, fmt.Errorf("can't unmarshal event due to: %v\n", err)
	}

	repositoryId, err := JsonResolveInt64(event, []string{"{repository,repo}", "id"})
	if err != nil {
		return RepositoryEvent{}, fmt.Errorf("can't extract repository id due to: %v\n", err)
	}

	eventType, err := JsonResolveString(event, []string{"type"})
	if err != nil {
		return RepositoryEvent{}, fmt.Errorf("can't extract event type due to: %v\n", err)
	}

	rawEventId, err := JsonResolveString(event, []string{"id"})
	if err != nil {
		return RepositoryEvent{}, fmt.Errorf("can't extract event id due to: %v\n", err)
	}

	eventId, err := strconv.ParseInt(rawEventId, 10, 64)
	if err != nil {
		return RepositoryEvent{}, fmt.Errorf("can't extract event id due to: %v\n", err)
	}

	return RepositoryEvent{
		RepositoryId: RepositoryId(repositoryId),
		EventType:    eventType,
		EventId:      RepositoryEventId(eventId),
		Event:        event,
	}, nil
}

func SaveRepositoryEvent(event RepositoryEvent, outPath string) error {
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outPath), os.ModePerm); err != nil {
		return err
	}
	if err := os.WriteFile(outPath, eventBytes, 0644); err != nil {
		return err
	}

	return nil
}

func SaveRepositoryEvents(events []RepositoryEvent, outPath string) error {
	eventsBytes, err := json.Marshal(events)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outPath), os.ModePerm); err != nil {
		return err
	}
	if err := os.WriteFile(outPath, eventsBytes, 0644); err != nil {
		return err
	}

	return nil
}

func LoadRepositoryEvent(inPath string) (RepositoryEvent, error) {
	eventBytes, err := os.ReadFile(inPath)
	if err != nil {
		return RepositoryEvent{}, err
	}

	var event RepositoryEvent
	err = json.Unmarshal(eventBytes, &event)
	if err != nil {
		return RepositoryEvent{}, err
	}

	return event, nil
}

func LoadRepositoryEvents(inPath string) ([]RepositoryEvent, error) {
	eventsBytes, err := os.ReadFile(inPath)
	if err != nil {
		return nil, err
	}

	events := make([]RepositoryEvent, 0)
	err = json.Unmarshal(eventsBytes, &events)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func LoadRepositoryEventsOrDefault(inPath string, defaultEvents []RepositoryEvent) ([]RepositoryEvent, error) {
	events, err := LoadRepositoryEvents(inPath)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return defaultEvents, nil
	default:
		return events, err
	}
}

func AggregateRepositoryEvents(inDirectory string, outDirectory string) error {
	matches, err := filepath.Glob(filepath.Join(inDirectory, "*.json"))
	if err != nil {
		return err
	}

	repositoryEventsByRepositoryId := make(map[RepositoryId][]RepositoryEvent)

	for _, match := range matches {
		event, err := LoadRepositoryEvent(match)
		if err != nil {
			return err
		}

		events, ok := repositoryEventsByRepositoryId[event.RepositoryId]
		if !ok {
			events = make([]RepositoryEvent, 0)
		}

		events = append(events, event)

		repositoryEventsByRepositoryId[event.RepositoryId] = events
	}

	for repositoryId, events := range repositoryEventsByRepositoryId {
		outPath := path.Join(outDirectory, fmt.Sprintf("%d.json", repositoryId))

		if err := SaveRepositoryEvents(events, outPath); err != nil {
			return err
		}
	}

	return nil
}

func (re *RepositoryEvent) Texts() []string {
	texts := make([]string, 0, 15)

	switch re.EventType {
	case "CommitCommentEvent":
		if text, err := JsonResolveString(re.Event, []string{"payload", "comment", "body"}); err == nil {
			texts = append(texts, text)
		}
	case "CreateEvent":
		if text, err := JsonResolveString(re.Event, []string{"payload", "description"}); err == nil {
			texts = append(texts, text)
		}
	case "IssueCommentEvent":
		if text, err := JsonResolveString(re.Event, []string{"payload", "comment", "body"}); err == nil {
			texts = append(texts, text)
		}
	case "IssuesEvent":
		if text, err := JsonResolveString(re.Event, []string{"payload", "issue", "body"}); err == nil {
			texts = append(texts, text)
		}
		if text, err := JsonResolveString(re.Event, []string{"payload", "issue", "title"}); err == nil {
			texts = append(texts, text)
		}
	case "PullRequestEvent":
		if text, err := JsonResolveString(re.Event, []string{"payload", "pull_request", "body"}); err == nil {
			texts = append(texts, text)
		}
		if text, err := JsonResolveString(re.Event, []string{"payload", "pull_request", "title"}); err == nil {
			texts = append(texts, text)
		}
	case "PullRequestReviewEvent":
		if text, err := JsonResolveString(re.Event, []string{"payload", "review", "body"}); err == nil {
			texts = append(texts, text)
		}
	case "PullRequestReviewCommentEvent":
		if text, err := JsonResolveString(re.Event, []string{"payload", "comment", "body"}); err == nil {
			texts = append(texts, text)
		}
	case "PushEvent":
		if commits, err := JsonResolveArray(re.Event, []string{"payload", "commits"}); err == nil {
			for _, commit := range commits {
				if text, err := JsonResolveString(commit, []string{"message"}); err == nil {
					texts = append(texts, text)
				}
			}
		}
	case "ReleaseEvent":
		if text, err := JsonResolveString(re.Event, []string{"payload", "release", "body"}); err == nil {
			texts = append(texts, text)
		}
		if text, err := JsonResolveString(re.Event, []string{"payload", "release", "name"}); err == nil {
			texts = append(texts, text)
		}
	// case "DeleteEvent":
	// case "ForkEvent":
	// case "GollumEvent":
	// case "MemberEvent":
	// case "PublicEvent":
	// case "PullRequestReviewThreadEvent":
	// case "SponsorshipEvent":
	// case "WatchEvent":
	default:
	}

	return texts
}

func (re *RepositoryEvent) CountKeywordMatches(keywords []string) int {
	texts := re.Texts()
	numKeywordMatches := 0

	for _, text := range texts {
		for _, keyword := range keywords {
			if strings.Contains(strings.ToLower(text), strings.ToLower(keyword)) {
				numKeywordMatches++
			}
		}
	}

	return numKeywordMatches
}

func ghArchiveParse(repositoryIds []RepositoryId, keywords []string, zippedBytes []byte) []RepositoryEvent {
	zippedBytesReader := bytes.NewReader(zippedBytes)
	bytesReader, err := gzip.NewReader(zippedBytesReader)
	if err != nil {
		fmt.Printf("Can't create gzip reader due to: %v\n", err)
		return []RepositoryEvent{}
	}

	events := make([]RepositoryEvent, 0, 300)

	scanner := bufio.NewScanner(bytesReader)
	for scanner.Scan() {
		rawEvent := scanner.Bytes()
		event, err := ParseRepositoryEvent(rawEvent)
		if err != nil {
			fmt.Printf("Can't parse event due to: %v\n", err)
			continue
		}

		if !ContainsRepositoryId(repositoryIds, event.RepositoryId) {
			continue
		}

		if event.CountKeywordMatches(keywords) <= 0 {
			continue
		}

		events = append(events, event)
	}

	return events
}

func GetRepositoryEventsForHour(
	httpClient *http.Client,
	repositoryIds []RepositoryId,
	keywords []string,
	date Date,
	hour int,
) []RepositoryEvent {
	repositoryEvents, err := RetryWithResult(func() ([]RepositoryEvent, error) {
		responseBytes, err := DownloadFile(
			httpClient,
			fmt.Sprintf("https://data.gharchive.org/%04d-%02d-%02d-%d.json.gz",
				date.Year,
				date.Month,
				date.Day,
				hour,
			),
		)

		if err != nil {
			return nil, fmt.Errorf("can't download repository events due to: %v\n", err)
		}

		return ghArchiveParse(repositoryIds, keywords, responseBytes), nil
	}, DefaultErrorHandler)

	if err != nil {
		fmt.Printf("can't get repository events for %s at %d due to: %v\n", date.ToString(), hour, err)
		return []RepositoryEvent{}
	}

	return repositoryEvents
}

func GetRepositoryEventsForDay(
	httpClient *http.Client,
	repositoryIds []RepositoryId,
	keywords []string,
	date Date,
) []RepositoryEvent {
	events := make([]RepositoryEvent, 0)
	for hour := 0; hour < 24; hour++ {
		events = append(events, GetRepositoryEventsForHour(httpClient, repositoryIds, keywords, date, hour)...)
	}
	fmt.Printf("processed events for %s\n", date.ToString())
	return events
}

func DownloadRepositoryEventsReadWorker(
	repositoryIds []RepositoryId,
	keywords []string,
	workQueue chan Date,
	resultQueue chan []RepositoryEvent,
	wgWorker, wgWorkQueue *sync.WaitGroup,
) {
	httpClient := &http.Client{}
	defer wgWorker.Done()
	for date := range workQueue {
		repositoryEvents := GetRepositoryEventsForDay(httpClient, repositoryIds, keywords, date)

		resultQueue <- repositoryEvents

		wgWorkQueue.Done()
	}
}

func DownloadRepositoryEventsWriteWorker(
	outputDirectory string,
	resultQueue chan []RepositoryEvent,
	wgReceiver *sync.WaitGroup,
) {
	defer wgReceiver.Done()

	for result := range resultQueue {
		for _, event := range result {
			outputPath := path.Join(
				outputDirectory,
				fmt.Sprintf("%d-%d.json", event.RepositoryId, event.EventId),
			)
			if err := SaveRepositoryEvent(event, outputPath); err != nil {
				fmt.Printf("Failed to save repository %d event %d due to: %v\n", event.RepositoryId, event.EventId, err)
			}
		}
	}
}

func DownloadRepositoryEvents(
	numWorkers int,
	repositoryIds []RepositoryId,
	dateRange DateRange,
	keywords []string,
	outputDirectory string,
) {
	workQueue := make(chan Date, 10_000)
	resultQueue := make(chan []RepositoryEvent, 10_000)

	var wgWorker sync.WaitGroup
	var wgReceiver sync.WaitGroup
	var wgWorkQueue sync.WaitGroup

	for workerIndex := 0; workerIndex < numWorkers; workerIndex++ {
		wgWorker.Add(1)
		go DownloadRepositoryEventsReadWorker(
			repositoryIds,
			keywords, workQueue,
			resultQueue,
			&wgWorker, &wgWorkQueue,
		)
	}

	wgReceiver.Add(1)
	go DownloadRepositoryEventsWriteWorker(
		outputDirectory,
		resultQueue,
		&wgReceiver,
	)

	for _, date := range dateRange.ToList() {
		wgWorkQueue.Add(1)
		workQueue <- date
	}
	wgWorkQueue.Wait()

	close(workQueue)
	close(resultQueue)

	wgWorker.Wait()
	wgReceiver.Wait()
}
