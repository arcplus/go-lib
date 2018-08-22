package token

import (
	"testing"
)

func TestSignAndValidate(t *testing.T) {
	Check()
	c := Claims{
		Issuer:  "sso",
		Subject: "uuid",
	}
	tokenStr := c.Sign()
	t.Log(tokenStr)

	nc, err := Validate(tokenStr)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(nc)
}
