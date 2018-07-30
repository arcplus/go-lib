package token

import (
	"crypto/rsa"
	"io/ioutil"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var privateKey *rsa.PrivateKey
var publicKey *rsa.PublicKey

func init() {
	priKeyStr := `-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQC/vqutzs6IRZfLcv1fJhJzDfzFuFLk6E1RhCUVKTQNUXjXXNYo
13Q9lFumlOhtpV6tdiYFkgShnoxEO5Cr4Oh11TnItokXTMaU2vGXC9jiCA7NupPe
usI+HyfrVlxD1P2n9XMiwn/TXxJiID/Wp8cOQ4+HPDG+Lyo1Gg2KEzM41wIDAQAB
AoGAfD9uedLvrAgEg7YAjv5ZqCphKDH3rRMGvxK1ANBRRWwMtOkYcSCj1x9igEAv
mJU3E4niu2tSCvR1CeXbKjU0C7+4JN84FdOJY8rotlGPxdYshR23dCtW4H36Ia9J
g4EGXq+evdj5wAYthHBi1KQHayrX47psOYWi+9d5NXv3sBkCQQDf+EUC0qsOSSAA
Uw9xjCI6a551FSVT6qoIlHqHUIbeJXgsvuGxHyFveQI1ggLM8QwggrbQRTpXUIrV
Ura7W0r7AkEA2yqdohemzGMFUXfpX2Q+qtKHSbYZnPsunb+vCRQHN2tfiJRYEr+g
l9x9MlGghoTfnJ1WSDmg0lTWLmUcVnGi1QJBANRY11Vt17CbtDOajLHjYzBwiLQJ
cHK3sq6f1+wjdTt52w7Ri7obAeBmoqmIso8Mm6rXQ+0DNeVC/95xpb7NN7ECQF60
66k/zzRDFek+h/pQt0PZ9dxEdI0BfgNs8ZZasUOhgobik6yGYj89aFx2KYf3oylq
U/6h6Hz7bBJgXv573IECQCLkAIvBrv+T548NQy/Xa+eAZAC1rCSkqH8hWRfUelke
lukTT9zN4Z89dv7Oo3exyukHDi5bN7MZXixNrHh5CDk=
-----END RSA PRIVATE KEY-----`

	pubKeyStr := `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC/vqutzs6IRZfLcv1fJhJzDfzF
uFLk6E1RhCUVKTQNUXjXXNYo13Q9lFumlOhtpV6tdiYFkgShnoxEO5Cr4Oh11TnI
tokXTMaU2vGXC9jiCA7NupPeusI+HyfrVlxD1P2n9XMiwn/TXxJiID/Wp8cOQ4+H
PDG+Lyo1Gg2KEzM41wIDAQAB
-----END PUBLIC KEY-----`

	// set default key
	SetPrivateKey(priKeyStr)
	SetPublickKey(pubKeyStr)
}

// SetPrivateKey set private key
func SetPrivateKey(key string) {
	var err error
	privateKey, err = jwt.ParseRSAPrivateKeyFromPEM([]byte(key))
	if err != nil {
		panic(err)
	}
}

// SetPublickKey set publick key
func SetPublickKey(key string) {
	var err error
	publicKey, err = jwt.ParseRSAPublicKeyFromPEM([]byte(key))
	if err != nil {
		panic(err)
	}
}

// SetKeyFromFile read pri/pub key from given file
func SetKeyFromFile(priName, pubName string) {
	pri, err := ioutil.ReadFile(priName)
	if err != nil {
		panic(err)
	}
	SetPrivateKey(string(pri))
	pub, err := ioutil.ReadFile(pubName)
	if err != nil {
		panic(err)
	}
	SetPublickKey(string(pub))
}

// Check checks if privateKey and publicKey are ok
// this should run after SetPrivateKey/SetPublickKey
func Check() {
	c := Claims{}
	_, err := Validate(c.Sign())
	if err != nil {
		panic(err)
	}
}

// Sign generate jwt token str
func (c *Claims) Sign() string {
	c.Version = Version
	c.IssuedAt = time.Now().Unix()

	tokenStr, err := jwt.NewWithClaims(jwt.SigningMethodRS256, c).SignedString(privateKey)
	if err != nil {
		panic(err)
	}
	return tokenStr
}

// Validate convert token str to *Claims if possible
func Validate(tokenStr string) (*Claims, error) {
	c := &Claims{}
	_, err := jwt.ParseWithClaims(tokenStr, c, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	// logical error
	if c.Version != Version {
		return nil, ErrVersionInvalid
	}

	if time.Now().Sub(time.Unix(c.IssuedAt, 0)) > Expire {
		if time.Now().Sub(time.Unix(c.IssuedAt, 0)) < MaxExpire {
			return c, ErrNeedRefresh
		}
		return nil, ErrExpired
	}

	return c, nil
}
