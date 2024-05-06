package xtypes

import (
	"strconv"

	"github.com/txix-open/isp-kit/json"
)

type Json[T any] struct {
	Value T
}

func (l Json[T]) MarshalJSON() ([]byte, error) {
	data, err := l.MarshalText()
	if err != nil {
		return nil, err
	}
	return []byte(strconv.Quote(string(data))), nil
}

func (l *Json[T]) UnmarshalJSON(data []byte) error {
	s, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}
	return l.UnmarshalText([]byte(s))
}

func (l Json[T]) MarshalText() ([]byte, error) {
	data, err := json.Marshal(l.Value)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (l *Json[T]) UnmarshalText(data []byte) error {
	return json.Unmarshal(data, &l.Value)
}
