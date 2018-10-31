package crypto

import (
	"math/rand"
	"strings"
	"time"
)

// ref: https://github.com/elgs/gostrgen

// set
const (
	Lower = 1 << 0
	Upper = 1 << 1
	Digit = 1 << 2
	Punct = 1 << 3

	LowerUpper      = Lower | Upper
	LowerDigit      = Lower | Digit
	UpperDigit      = Upper | Digit
	LowerUpperDigit = LowerUpper | Digit
	All             = LowerUpperDigit | Punct
)

const lower = "abcdefghijklmnopqrstuvwxyz"
const upper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const digit = "0123456789"
const punct = "~!@#$%^&*()_+-="

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// RandString default set LowerDigit
func RandString(size int, sets ...int) string {
	return RandStringExcl(size, "", sets...)
}

// RandNum default set 0123456789
func RandNumStr(size int, excludes ...string) string {
	var excl string
	for i := range excludes {
		excl += excludes[i]
	}
	return RandStringExcl(size, excl, Digit)
}

func RandStringExcl(size int, exclude string, sets ...int) string {
	var set int

	for i := range sets {
		set = sets[i] | set
	}

	if set == 0 {
		set = LowerDigit
	}

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
	if set&Punct > 0 {
		charset += punct
	}

	lenAll := len(charset)
	if len(exclude) >= lenAll {
		panic("Too much to exclude.")
	}

	buf := make([]byte, size)
	for i := 0; i < size; i++ {
		b := charset[rand.Intn(lenAll)]
		if exclude != "" {
			if strings.Contains(exclude, string(b)) {
				i--
				continue
			}
		}
		buf[i] = b
	}
	return string(buf)
}
