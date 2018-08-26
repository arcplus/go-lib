package micro

import (
	"errors"
	"testing"
)

func TestVersionInfo(t *testing.T) {
	version = "1.0"
	gitCommit = "d12e63e8"
	buildDate = "2018-08-26"
	t.Logf("\n%s", VersionInfo())
}

func TestMicro_Close(t *testing.T) {
	m := New()
	m.Close()

	m.AddResCloseFunc(func() error {
		return errors.New("res1")
	})

	m.AddResCloseFunc(func() error {
		return errors.New("res2")
	})

	m.Close()
	m.Close()
}
