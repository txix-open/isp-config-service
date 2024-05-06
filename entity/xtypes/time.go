package xtypes

import (
	"strconv"
	"time"
)

type Time struct {
	Value time.Time
}

func (l Time) MarshalJSON() ([]byte, error) {
	return l.MarshalText()
}

func (l *Time) UnmarshalJSON(data []byte) error {
	return l.UnmarshalText(data)
}

func (l Time) MarshalText() ([]byte, error) {
	value := strconv.Itoa(int(l.Value.Unix()))
	return []byte(value), nil
}

func (l *Time) UnmarshalText(data []byte) error {
	value, err := strconv.Atoi(string(data))
	if err != nil {
		return err
	}
	l.Value = time.Unix(int64(value), 0)
	return nil
}
