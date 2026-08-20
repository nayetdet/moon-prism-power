package utils

import "fmt"

func MediaKey(kind string, id int) string {
	return fmt.Sprintf("%s:%d", kind, id)
}
