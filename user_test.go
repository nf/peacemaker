package main

import (
	"testing"
	"time"
)

var debitTestUser = &User{
	Balance: []*Balance{
		{"anytime", 1, 0},
		{"weekday", 1, 1},
		{"weekend", 1, 2},
	},
}

var debitTests = []struct {
	t   time.Time
	n   int
	err error
}{
	// Thursday 10am (should deduct anytime minutes)
	{time.Date(2012, time.May, 10, 10, 0, 0, 0, time.Local), 1, nil},
	{time.Date(2012, time.May, 10, 10, 0, 0, 0, time.Local), 1, ZeroBalance},

	// Thursday 5pm (should deduct weekday minutes)
	{time.Date(2012, time.May, 10, 17, 0, 0, 0, time.Local), 1, nil},
	{time.Date(2012, time.May, 10, 17, 0, 0, 0, time.Local), 1, ZeroBalance},

	// Friday 5pm (should deduct weekend minutes)
	{time.Date(2012, time.May, 11, 17, 0, 0, 0, time.Local), 1, nil},
	{time.Date(2012, time.May, 11, 17, 0, 0, 0, time.Local), 1, ZeroBalance},
}

func TestDebit(t *testing.T) {
	for i, test := range debitTests {
		err := debitTestUser.Debit(test.t, test.n)
		if err != test.err {
			t.Errorf("%d: got err = %v, want %v", i, err, test.err)
		}
	}
}
