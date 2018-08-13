package seq

import (
	"strconv"
	"time"

	"github.com/sony/sonyflake"
)

var sf *sonyflake.Sonyflake

func init() {
	var st sonyflake.Settings
	st.StartTime = time.Date(2016, 9, 1, 0, 0, 0, 0, time.UTC)
	sf = sonyflake.NewSonyflake(st)
	if sf == nil {
		panic("sonyflake not created")
	}
}

// NextID return an unique id str
func NextID() string {
	id, _ := sf.NextID()
	return strconv.FormatUint(id, 10)
}

// NextNumID returns number id
func NextNumID() uint64 {
	id, _ := sf.NextID()
	return id
}
