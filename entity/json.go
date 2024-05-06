package entity

import (
	"encoding/base64"
	"strconv"

	"github.com/txix-open/isp-kit/json"
)

type JsonValue[T any] struct {
	Value T
}

func (l JsonValue[T]) MarshalJSON() ([]byte, error) {
	data, err := l.MarshalText()
	if err != nil {
		return nil, err
	}
	return []byte(strconv.Quote(string(data))), nil
}

func (l *JsonValue[T]) UnmarshalJSON(data []byte) error {
	s, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}
	return l.UnmarshalText([]byte(s))
}

func (l JsonValue[T]) MarshalText() ([]byte, error) {
	data, err := json.Marshal(l.Value)
	if err != nil {
		return nil, err
	}
	s := base64.StdEncoding.EncodeToString(data)
	return []byte(s), nil
}

func (l *JsonValue[T]) UnmarshalText(data []byte) error {
	arr, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return err
	}
	return json.Unmarshal(arr, &l.Value)
}
