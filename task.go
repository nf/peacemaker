package main

import "time"

type Refill struct {
	RunAt   time.Time
	Period  time.Duration
	Balance string
	Minutes int
}

func (r *Refill) Do(now time.Time, user map[string]*User) bool {
	if now.Before(r.RunAt) {
		return false
	}
	r.RunAt = r.RunAt.Add(r.Period)
	for _, u := range user {
		u.SetBalance(r.Balance, r.Minutes)
	}
	return true
}
