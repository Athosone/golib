package utils

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// Generates a random name with a prefix
// in the format of: prefix-randomstring
func GenerateRandomNameWithPrefix(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, RandomString(5))
}

func RandomString(len int) string {
	var b strings.Builder
	rand.Seed(time.Now().UnixNano())

	b.Grow(len)
	for i := 0; i < len; i++ {
		b.WriteByte(byte(65 + rand.Intn(25)))
	}
	return strings.ToLower(b.String())
}
