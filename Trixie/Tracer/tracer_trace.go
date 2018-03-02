// +build trace

package Tracer

import (
	"log"
	"os"
	"time"
)

var logger = log.New(os.Stderr, "[TRACE] ", 0)

type TrackType struct {
	name  string
	start time.Time
}

func Log(format string, v ...interface{}) {
	logger.Printf(format, v...)
}

func Track(name string) TrackType {
	start := time.Now()
	logger.Printf("[TRACK] '%s'\tSTART", name)
	return TrackType{name, start}
}

func Un(track TrackType) {
	logger.Printf("[TRACK] '%s'\t-> %s", track.name, time.Since(track.start))
}
