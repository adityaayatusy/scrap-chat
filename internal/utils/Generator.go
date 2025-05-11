package utils

import (
	"math/rand"
	"time"
)

func GenerateZX() string {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	letters := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, 16)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}
