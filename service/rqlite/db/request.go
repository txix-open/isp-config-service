package db

func request(query string, args ...any) []any {
	return append([]any{query}, args...)
}

func requests(requests ...[]any) [][]any {
	return requests
}
