package models

import (
	"fmt"
	"time"
)

const TimeFormat = time.RFC3339

type CustomTime struct {
	time.Time
}

func (t *CustomTime) MarshalJSON() ([]byte, error) {
	if t.Time.IsZero() {
		return nil, nil
	}
	timeString := fmt.Sprintf("\"%s\"", t.Time.Format(TimeFormat))
	return []byte(timeString), nil
}
