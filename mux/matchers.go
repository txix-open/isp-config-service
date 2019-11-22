package mux

import (
	"bytes"
)

type Matcher func([]byte) bool

// Any is a Matcher that matches any connection.
func Any() Matcher {
	return func(_ []byte) bool {
		return true
	}
}

var httpMethods = map[string]struct{}{
	"OPTIONS": {},
	"HEAD":    {},
	"GET":     {},
	"POST":    {},
	"PUT":     {},
	"PATCH":   {},
	"DELETE":  {},
	"TRACE":   {},
	"CONNECT": {},
}

// HTTP1 only matches the methods in the HTTP request.
func HTTP1() Matcher {
	return func(data []byte) bool {
		// max length of method
		l := 7
		if len(data) < l {
			l = len(data)
		}
		methodB := data[:l]
		words := bytes.Split(methodB, []byte(" "))
		if len(words) == 0 {
			return false
		}
		_, ok := httpMethods[string(words[0])]
		return ok
	}
}
