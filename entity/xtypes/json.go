package xtypes

import (
	"strconv"

	"github.com/pkg/errors"
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
		return errors.WithMessage(err, "unquote text")
	}
	return l.UnmarshalText([]byte(s))
}

func (l Json[T]) MarshalText() ([]byte, error) {
	data, err := json.Marshal(l.Value)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal json")
	}
	return data, nil
}

func (l *Json[T]) UnmarshalText(data []byte) error {
	err := json.Unmarshal(data, &l.Value)
	if err != nil {
		return errors.WithMessage(err, "unmarshal json")
	}
	return nil
}
