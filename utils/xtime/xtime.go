package xtime

import (
	"time"
)

var now func() time.Time = time.Now

func Mock(mockedTime time.Time) {
	now = func() time.Time {
		return mockedTime
	}
}

func Unmock() {
	now = time.Now
}

func Now() time.Time {
	return now()
}
