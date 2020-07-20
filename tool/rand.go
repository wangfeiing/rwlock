package tool

import (
	"math/rand"
	"time"
)

func Rand(min, max int) int {
	return rand.New(rand.NewSource(time.Now().UnixNano())).Intn(max-min) + min
}
