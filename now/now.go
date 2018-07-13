package now

import (
	"strconv"
	"time"
)

type Cur struct {
	time.Time
}

func Now() *Cur {
	return &Cur{time.Now()}
}

func New(t time.Time) *Cur {
	return &Cur{t}
}

func NewUnix(sec interface{}) *Cur {
	switch v := sec.(type) {
	case int64:
		return &Cur{time.Unix(v, 0)}
	case int:
		return &Cur{time.Unix(int64(v), 0)}
	case string:
		s, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil
		}
		return &Cur{time.Unix(s, 0)}
	}

	return nil
}

func (c *Cur) BeginningOfHour() time.Time {
	return c.Truncate(time.Hour)
}

func BeginningOfHour() time.Time {
	return New(time.Now()).BeginningOfHour()
}

func (c *Cur) BeginningOfDay() time.Time {
	d := time.Duration(-c.Hour()) * time.Hour
	return c.Truncate(time.Hour).Add(d)
}

func BeginningOfDay() time.Time {
	return New(time.Now()).BeginningOfDay()
}

func (c *Cur) BeginningOfWeek(mFirst bool) time.Time {
	t := c.BeginningOfDay()
	weekday := int(t.Weekday())
	if mFirst {
		if weekday == 0 {
			weekday = 7
		}
		weekday = weekday - 1
	}

	d := time.Duration(-weekday) * 24 * time.Hour
	return t.Add(d)
}

func BeginningOfWeek(mFirst bool) time.Time {
	return New(time.Now()).BeginningOfWeek(mFirst)
}

func (c *Cur) BeginningOfMonth() time.Time {
	t := c.BeginningOfDay()
	d := time.Duration(-int(t.Day())+1) * 24 * time.Hour
	return t.Add(d)
}

func BeginningOfMonth() time.Time {
	return New(time.Now()).BeginningOfMonth()
}

func (c *Cur) BeginningOfQuarter() time.Time {
	month := c.BeginningOfMonth()
	offset := (int(month.Month()) - 1) % 3
	return month.AddDate(0, -offset, 0)
}

func BeginningOfQuarter() time.Time {
	return New(time.Now()).BeginningOfQuarter()
}

func (c *Cur) EndOfHour() time.Time {
	return c.BeginningOfHour().Add(time.Hour - time.Nanosecond)
}

func EndOfHour() time.Time {
	return New(time.Now()).EndOfHour()
}

func (c *Cur) EndOfDay() time.Time {
	return c.BeginningOfDay().Add(24*time.Hour - time.Nanosecond)
}

func EndOfDay() time.Time {
	return New(time.Now()).EndOfDay()
}

func (c *Cur) EndOfWeek(mFirst bool) time.Time {
	return c.BeginningOfWeek(mFirst).AddDate(0, 0, 7).Add(-time.Nanosecond)
}

func EndOfWeek(mFirst bool) time.Time {
	return New(time.Now()).EndOfWeek(mFirst)
}

func (c *Cur) EndOfMonth() time.Time {
	return c.BeginningOfMonth().AddDate(0, 1, 0).Add(-time.Nanosecond)
}

func EndOfMonth() time.Time {
	return New(time.Now()).EndOfMonth()
}

func (c *Cur) EndOfQuarter() time.Time {
	return c.BeginningOfQuarter().AddDate(0, 3, 0).Add(-time.Nanosecond)
}

func EndOfQuarter() time.Time {
	return New(time.Now()).EndOfQuarter()
}

func (c *Cur) Format(f string) string {
	return c.Time.Format(f)
}

func Format(f string) string {
	return New(time.Now()).Format(f)
}

func Parse(layout string, value string) (*Cur, error) {
	t, err := time.Parse(layout, value)
	return &Cur{t}, err
}

func ParseInLocation(layout string, value string, loc *time.Location) (*Cur, error) {
	if loc == nil {
		loc = time.Local
	}
	t, err := time.ParseInLocation(layout, value, loc)
	return &Cur{t}, err
}
