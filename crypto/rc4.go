package crypto

import "crypto/rc4"

// RC4 rc4 xor
func RC4(src, key []byte) ([]byte, error) {
	cipher, err := rc4.NewCipher(key)
	if err != nil {
		return nil, err
	}

	dst := make([]byte, len(src))
	cipher.XORKeyStream(dst, src)
	return dst, nil
}
