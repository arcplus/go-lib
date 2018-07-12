package crypto

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"hash/crc32"
)

// Sha1Str returns sha1 string
func Sha1Str(str string) string {
	h := sha1.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// Md5Str returns md5 str
func Md5Str(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// Crc32 returns crc32 int
func Crc32(str string) uint32 {
	return crc32.ChecksumIEEE([]byte(str))
}
