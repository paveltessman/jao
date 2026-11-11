package dbtools

import (
	"errors"
	"time"
)

var ErrDb = errors.New("db error")

func Int64ToUnixTime(val int64) time.Time {
	return time.Unix(val, 0)
}

func TimeToInt64(val time.Time) int64 {
	return val.Unix()
}
