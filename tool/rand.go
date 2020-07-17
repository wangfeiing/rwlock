package tool

import (
	"math/rand"
	"time"
)

var genRandom = rand.New(rand.NewSource(time.Now().UnixNano()))

func Rand(min, max int) int {
	return genRandom.Intn(max-min) + min
}
