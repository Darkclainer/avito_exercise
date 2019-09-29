package main

import (
	"database/sql/driver"
	"time"
)

// A MatchTime helps to match sqlmock parameter between two moment of time.
// While create a MatchTime you must specify Start.
// You can specify End, but if you not - while matching will be used value time.Now()
type MatchTime struct {
	Start time.Time
	End   time.Time
}

func (matchTime MatchTime) Match(value driver.Value) bool {
	timeValue, ok := value.(time.Time)
	if !ok {
		return false
	}
	if matchTime.End.IsZero() {
		matchTime.End = time.Now()
	}
	if timeValue.Before(matchTime.Start) || timeValue.After(matchTime.End) {
		return false
	}
	return true
}
