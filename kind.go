package main

import "time"

var timeZone = time.Local

type Kind interface {
	Available(time.Time) bool
}

var kindMap = map[string]Kind{
	"weekday": Weekday{},
	"weekend": Weekend{},
}

type Weekday struct{}

func (k Weekday) Available(t time.Time) bool {
	t = t.In(timeZone)

	// Can use weekday credit on weekends.
	if d := t.Weekday(); d == time.Saturday || d == time.Sunday {
		return true
	}

	// Can only use weekday credit after 5pm.
	return t.Hour() >= 17
}

type Weekend struct{}

func (k Weekend) Available(t time.Time) bool {
	t = t.In(timeZone)

	// Weekend starts at 5pm Friday and ends Sunday night.
	d := t.Weekday()
	if d == time.Friday && t.Hour() >= 17 {
		return true
	}
	return d == time.Saturday || d == time.Sunday
}
