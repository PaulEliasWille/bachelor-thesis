package main

import (
	"fmt"
	"time"
)

type Date struct {
	Year  int
	Month int
	Day   int
}

func (d Date) ToString() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}

func (d Date) IsBefore(other Date) bool {
	if d.Year < other.Year {
		return true
	}
	if d.Year > other.Year {
		return false
	}
	if d.Month < other.Month {
		return true
	}
	if d.Month > other.Month {
		return false
	}
	return d.Day < other.Day
}

func (d Date) IsEqual(other Date) bool {
	return d.Year == other.Year && d.Month == other.Month && d.Day == other.Day
}

func (d Date) IsAfter(other Date) bool {
	return !d.IsBefore(other) && !d.IsEqual(other)
}

func (d Date) AddDays(numDays int) Date {
	resultTime := time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.UTC).AddDate(0, 0, numDays)
	return Date{
		Year:  resultTime.Year(),
		Month: int(resultTime.Month()),
		Day:   resultTime.Day(),
	}
}

func (d Date) DaysUntil(other Date) int {
	time1 := time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.UTC)
	time2 := time.Date(other.Year, time.Month(other.Month), other.Day, 0, 0, 0, 0, time.UTC)
	duration := time2.Sub(time1)
	return int(duration.Hours() / 24)
}

func (d Date) Previous() Date {
	return d.AddDays(-1)
}

func (d Date) Next() Date {
	return d.AddDays(1)
}
