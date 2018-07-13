package now

import (
	"testing"
	"time"
)

func TestBeginningOfDay(t *testing.T) {
	t.Log(BeginningOfDay())
}

func TestBeginningOfWeek(t *testing.T) {
	t.Log(BeginningOfWeek(false))
	t.Log(BeginningOfWeek(true))
}

func TestBeginningOfMonth(t *testing.T) {
	t.Log(BeginningOfMonth())
}

func TestBeginningOfQuarter(t *testing.T) {
	t.Log(BeginningOfQuarter())
}

func TestEndOfDay(t *testing.T) {
	t.Log(EndOfDay())
}

func TestEndOfWeek(t *testing.T) {
	t.Log(EndOfWeek(false))
	t.Log(EndOfWeek(true))
}

func TestEndOfMonth(t *testing.T) {
	t.Log(EndOfMonth())
}

func TestEndOfQuarter(t *testing.T) {
	t.Log(EndOfQuarter())
}

func TestNewUnix(t *testing.T) {
	t.Log(NewUnix(1486537851).Format("2006-01-02"))
}

func TestParse(t *testing.T) {
	WXPAYTIMESTAMP := "20060102150405"
	cur := Now().Truncate(time.Second)

	str := cur.Add(time.Minute * 30).Format(WXPAYTIMESTAMP)
	t.Log(str)

	pt, err := Parse(WXPAYTIMESTAMP, str)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(pt)

	t.Log(pt.Sub(cur))
}
