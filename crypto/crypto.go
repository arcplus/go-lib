package crypto

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"hash/crc32"

	"golang.org/x/crypto/pbkdf2"
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

// HashPassword gen hashed password with given salt.
func HashPassword(passwd string, salt string) string {
	tempPasswd := pbkdf2.Key([]byte(passwd), []byte(salt), 2048, 32, sha256.New)
	return fmt.Sprintf("%x", tempPasswd)
}

// ConstantTimeCompare is used for password comparison in constant time.
func ConstantTimeCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
