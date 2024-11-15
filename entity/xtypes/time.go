package xtypes

import (
	"strconv"
	"time"

	"github.com/pkg/errors"
)

type Time time.Time

func (l Time) MarshalJSON() ([]byte, error) {
	return l.MarshalText()
}

func (l *Time) UnmarshalJSON(data []byte) error {
	return l.UnmarshalText(data)
}

func (l Time) MarshalText() ([]byte, error) {
	t := time.Time(l)
	value := strconv.Itoa(int(t.Unix()))
	return []byte(value), nil
}

func (l *Time) UnmarshalText(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	value, err := strconv.Atoi(string(data))
	if err != nil {
		return errors.WithMessagef(err, "parse int: %s", string(data))
	}
	*l = Time(time.Unix(int64(value), 0))
	return nil
}
