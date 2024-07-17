package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func NewHTTPClient(proxyFilePath string, proxyIndex int) *http.Client {
	file, err := os.Open(proxyFilePath)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	proxyUrls := make([]string, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lineParts := strings.FieldsFunc(line, func(r rune) bool {
			return r == ':'
		})
		if len(lineParts) != 4 {
			panic(err)
		}
		proxyIp := lineParts[0]
		proxyPort := lineParts[1]
		proxyUser := lineParts[2]
		proxyPassword := lineParts[3]

		proxyUrls = append(proxyUrls, fmt.Sprintf("http://%s:%s@%s:%s", proxyUser, proxyPassword, proxyIp, proxyPort))
	}

	safeProxyIndex := proxyIndex % len(proxyUrls)
	proxyUrl, err := url.Parse(proxyUrls[safeProxyIndex])
	if err != nil {
		panic(err)
	}

	return &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
}

func ParseRetryAfterHeader(retryAfter string) (time.Duration, error) {
	if dur, err := parseSeconds(retryAfter); err == nil {
		return dur, nil
	}
	if dt, err := parseHTTPDate(retryAfter); err == nil {
		return time.Until(dt), nil
	}
	return 0, errors.New("retry-After value must be seconds integer or HTTP date string")
}

func parseSeconds(retryAfter string) (time.Duration, error) {
	seconds, err := strconv.ParseInt(retryAfter, 10, 64)
	if err != nil {
		return time.Duration(0), err
	}
	if seconds < 0 {
		return time.Duration(0), errors.New("negative seconds not allowed")
	}
	return time.Second * time.Duration(seconds), nil
}

func parseHTTPDate(retryAfter string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC1123, retryAfter)
	if err != nil {
		return time.Time{}, err
	}
	return parsed, nil
}

func DownloadFile(httpClient *http.Client, url string) ([]byte, error) {
	response, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(response.Body)

	if response.StatusCode == http.StatusTooManyRequests {
		if retryAfter, err := ParseRetryAfterHeader(response.Header.Get("Retry-After")); err == nil {
			return nil, &ThrottledError{retryAfter: retryAfter}
		}
		return nil, &ThrottledError{retryAfter: 60 * time.Second}
	} else if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return responseBytes, nil
}

func DownloadJson(httpClient *http.Client, url string) (interface{}, error) {
	bytes, err := DownloadFile(httpClient, url)
	if err != nil {
		return nil, err
	}

	var result interface{}
	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func SimpleDownloadFile(url string) ([]byte, error) {
	httpClient := http.Client{}
	return DownloadFile(&httpClient, url)
}

func SimpleDownloadJson(url string) (interface{}, error) {
	httpClient := http.Client{}
	return DownloadJson(&httpClient, url)
}
