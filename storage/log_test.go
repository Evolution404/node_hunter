package storage

import (
	"strconv"
	"testing"
)

func TestWrite(t *testing.T) {
	l := StartLog()
	for i := 0; i < 10000; i++ {
		l.Relation <- strconv.Itoa(i)
	}
	l.Close()
}
