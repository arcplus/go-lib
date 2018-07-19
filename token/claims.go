package token

import (
	"errors"
	"time"
)

var Version = "1.0"
var Exipre = time.Hour * 2
var MaxExpire = time.Hour * 24 * 30

type Claims struct {
	Issuer   string `json:"iss,omitempty"`
	Subject  string `json:"sub,omitempty"` // user id
	Audience string `json:"aud,omitempty"`
	IssuedAt int64  `json:"iat,omitempty"`
	JwtId    string `json:"jti,omitempty"`
	Version  string `json:"ver,omitempty"` // private version
	Session  string `json:"ses,omitempty"` // private session
}

var ErrVersionInvalid = errors.New("token version has changed")
var ErrNeedRefresh = errors.New("token need refresh")
var ErrExpired = errors.New("token expired")

// Valid implement jwt Claims interface
func (c Claims) Valid() error {
	if c.Version != Version {
		return ErrVersionInvalid
	}

	if time.Now().Sub(time.Unix(c.IssuedAt, 0)) > Exipre {
		if time.Now().Sub(time.Unix(c.IssuedAt, 0)) < MaxExpire {
			return ErrNeedRefresh
		}
		return ErrExpired
	}

	return nil
}
