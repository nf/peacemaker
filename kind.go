package main

import "time"

var timeZone = time.Local

type Kind interface {
	Available(time.Time) bool
	Times() string
}

var kindMap = map[string]Kind{
	"anytime": Anytime{},
	"weekday": Weekday{},
	"weekend": Weekend{},
}

type Anytime struct{}

func (k Anytime) Available(_ time.Time) bool {
	return true
}

func (Anytime) Times() string {
	return "any time"
}

type Weekday struct{}

func (k Weekday) Available(t time.Time) bool {
	t = t.In(timeZone)

	// Can use weekday credit on weekends.
	if d := t.Weekday(); d == time.Saturday || d == time.Sunday {
		return true
	}

	// Can only use weekday credit after 4pm.
	return t.Hour() >= 16
}

func (Weekday) Times() string {
	return "4pm to midnight, Monday to Thursday"
}

type Weekend struct{}

func (k Weekend) Available(t time.Time) bool {
	t = t.In(timeZone)

	// Weekend starts at 3pm Friday and ends Sunday night.
	d := t.Weekday()
	if d == time.Friday && t.Hour() >= 17 {
		return true
	}
	return d == time.Saturday || d == time.Sunday
}

func (Weekend) Times() string {
	return "5pm Friday to midnight Sunday"
}
