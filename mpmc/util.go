package mpmc

import (
	"crypto/rand"
	"fmt"
	"strings"
)

func CreateID() string {
	b := make([]byte, 16) // Generate 16 random bytes
	_, err := rand.Read(b)
	if err != nil {
		panic("Error: unable to generate random bytes")
	}

	// Convert random bytes to a UUID-like string
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4],   // 8 hex digits
		b[4:6],   // 4 hex digits
		b[6:8],   // 4 hex digits
		b[8:10],  // 4 hex digits
		b[10:16], // 12 hex digits
	)
}

func TypeName[T any]() string {
	temp := fmt.Sprintf("%T", new(T))
	temp = strings.TrimPrefix(temp, "*")

	if lastDot := strings.LastIndex(temp, "."); lastDot != -1 {
		temp = temp[lastDot+1:]
	}

	return temp
}

func equalSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
