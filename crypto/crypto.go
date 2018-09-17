package crypto

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"hash/crc32"
	"hash/crc64"

	"golang.org/x/crypto/pbkdf2"
)

// SHA1Sum
func SHA1Sum(str string) string {
	buf := sha1.Sum([]byte(str))
	return hex.EncodeToString(buf[:])
}

// SHA256Sum
func SHA256Sum(str string) string {
	buf := sha256.Sum256([]byte(str))
	return hex.EncodeToString(buf[:])
}

// Md5Str
func Md5Str(str string) string {
	buf := md5.Sum([]byte(str))
	return hex.EncodeToString(buf[:])
}

// Crc32 returns crc32 int
func Crc32(str string) uint32 {
	return crc32.ChecksumIEEE([]byte(str))
}

var tabECMA = crc64.MakeTable(crc64.ECMA)

func Crc64(str string) uint64 {
	hash := crc64.New(tabECMA)
	hash.Write([]byte(str))
	return hash.Sum64()
}

// HashPassword gen hashed password with given salt.
func HashPassword(passwd string, salt string) string {
	tempPasswd := pbkdf2.Key([]byte(passwd), []byte(salt), 2048, 32, sha256.New)
	return hex.EncodeToString(tempPasswd)
}

// ConstantTimeCompare is used for password comparison in constant time.
func ConstantTimeCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
