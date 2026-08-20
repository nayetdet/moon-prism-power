package utils

import "io"

const maxErrorBodySize = 8 << 10

func ReadErrorBody(body io.Reader) string {
	data, _ := io.ReadAll(io.LimitReader(body, maxErrorBodySize))
	return string(data)
}
