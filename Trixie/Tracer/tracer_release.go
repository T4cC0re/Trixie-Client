// +build !trace

package Tracer

import (
	"time"
)

type TrackType struct {
	name  string
	start time.Time
}

var dummyTime = time.Now()

func Log(Pattern string, v ...interface{}) {
}

func Track(name string) TrackType {
	return TrackType{"", dummyTime}
}

func Un(track TrackType) {
}
