package seq

import (
	"strconv"
	"time"

	"github.com/sony/sonyflake"
)

var (
	startTime    = time.Date(2016, 9, 1, 0, 0, 0, 0, time.UTC)
	startTimeNum = toSonyflakeTime(startTime)
	sf           *sonyflake.Sonyflake
)

func init() {
	var st sonyflake.Settings
	st.StartTime = startTime
	sf = sonyflake.NewSonyflake(st)
	if sf == nil {
		panic("sonyflake not created")
	}
}

// NextNumID returns number id
func NextNumID() uint64 {
	id, _ := sf.NextID()
	return id
}

// NextID return an unique id str
func NextID() string {
	return strconv.FormatUint(NextNumID(), 10)
}

const sonyflakeTimeUnit = 1e7 // nsec, i.e. 10 msec

func toSonyflakeTime(t time.Time) int64 {
	return t.UTC().UnixNano() / sonyflakeTimeUnit
}

// NextNumIDByTime is used for gen seq for query [second accuracy]
func NextNumIDByTime(t time.Time) uint64 {
	elapsedTime := toSonyflakeTime(t) - startTimeNum
	id := uint64(elapsedTime)<<(sonyflake.BitLenSequence+sonyflake.BitLenMachineID) |
		uint64(0)<<sonyflake.BitLenMachineID |
		uint64(0)
	return id
}

// NextIDByTime is used for gen seq for query [second accuracy]
func NextIDByTime(t time.Time) string {
	return strconv.FormatUint(NextNumIDByTime(t), 10)
}
