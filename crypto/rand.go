package crypto

import (
	"math/rand"
	"time"
)

// set
const (
	Lower = 1 << 0
	Upper = 1 << 1
	Digit = 1 << 2

	LowerUpper      = Lower | Upper
	LowerDigit      = Lower | Digit
	UpperDigit      = Upper | Digit
	LowerUpperDigit = LowerUpper | Digit
)

const lower = "abcdefghijklmnopqrstuvwxyz"
const upper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const digit = "0123456789"

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// RandString RandString
func RandString(size int, set int) string {
	charset := ""
	if set&Lower > 0 {
		charset += lower
	}
	if set&Upper > 0 {
		charset += upper
	}
	if set&Digit > 0 {
		charset += digit
	}

	lenAll := len(charset)

	buf := make([]byte, size)
	for i := 0; i < size; i++ {
		buf[i] = charset[rand.Intn(lenAll)]
	}
	return string(buf)
}
