package mongostore

import "fmt"

func edgeIdFromPointIds(from, to int64) string {
	return fmt.Sprintf("%d-%d", min(from, to), max(from, to))
}
