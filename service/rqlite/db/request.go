package db

func Request(query string, args ...any) []any {
	return append([]any{query}, args...)
}
