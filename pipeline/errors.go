package main

import (
	"errors"
	"fmt"
	"time"
)

type ThrottledError struct {
	retryAfter time.Duration
}

func (err *ThrottledError) Error() string {
	return fmt.Sprintf("throttled, try again in %s", err.retryAfter.String())
}

type ErrorHandlerAction interface {
	ShouldRetry() bool
	BeforeRetry()
}

type ErrorHandlerActionRetryAfter struct {
	retryAfter time.Duration
}

func (e *ErrorHandlerActionRetryAfter) ShouldRetry() bool {
	return true
}

func (e *ErrorHandlerActionRetryAfter) BeforeRetry() {
	fmt.Printf("Retrying after %s", e.retryAfter.String())
	time.Sleep(e.retryAfter)
}

type ErrorHandlerActionDoNotRetry struct{}

func (e *ErrorHandlerActionDoNotRetry) ShouldRetry() bool {
	return false
}

func (e *ErrorHandlerActionDoNotRetry) BeforeRetry() {}

type ErrorHandler func(error) ErrorHandlerAction

var DefaultErrorHandler ErrorHandler = func(err error) ErrorHandlerAction {
	var throttledErr *ThrottledError
	switch {
	case errors.As(err, &throttledErr):
		return &ErrorHandlerActionRetryAfter{retryAfter: throttledErr.retryAfter}
	default:
		return &ErrorHandlerActionDoNotRetry{}
	}
}

func RetryWithoutResult(
	fn func() error,
	errorHandler ErrorHandler,
) error {
	for {
		err := fn()

		action := errorHandler(err)

		if action == nil || !action.ShouldRetry() {
			return err
		}

		action.BeforeRetry()
	}
}

func RetryWithResult[T any](
	fn func() (T, error),
	errorHandler ErrorHandler,
) (T, error) {
	var finalResult T
	finalErr := RetryWithoutResult(func() error {
		result, err := fn()
		finalResult = result
		return err
	}, errorHandler)
	return finalResult, finalErr
}
