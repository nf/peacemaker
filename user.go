package main

import (
	"errors"
	"sort"
	"time"
)

const (
	deadTimeout = time.Minute * 2
	gracePeriod = time.Minute * 1
)

var ZeroBalance = errors.New("no minutes available")

type User struct {
	Balance   []*Balance
	StartTime time.Time
	LastSeen  time.Time
}

func (u *User) Start(t time.Time) error {
	u.StartTime = t
	return u.Debit(t, 1)
}

func (u *User) Stop() {
	u.StartTime = time.Time{}
	u.LastSeen = time.Time{}
}

func (u *User) Running(t time.Time) bool {
	return t.Sub(u.LastSeen) < deadTimeout
}

func (u *User) Debit(t time.Time, mins int) error {
	u.LastSeen = t
	// Don't deduct minutes during grace period.
	if t.Sub(u.StartTime) < gracePeriod {
		return nil
	}
	for mins > 0 {
		b := u.AvailableBalance(t)
		if b == nil {
			return ZeroBalance
		}
		n, err := b.Debit(t, mins)
		mins -= n
		if err == ZeroBalance && n > 0 {
			if mins == 0 {
				// Exactly drained this balance.
				break
			}
			// Try next balance.
			continue
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *User) AvailableBalance(t time.Time) *Balance {
	sort.Sort(balanceSlice(u.Balance)) // sort by Priority
	for _, b := range u.Balance {
		if b.Available(t) {
			return b
		}
	}
	return nil
}

type Balance struct {
	Kind     string
	Minutes  int
	Priority int // balances with lower priorities will be drained first
}

func (b *Balance) Available(t time.Time) bool {
	k := kindMap[b.Kind]
	if k == nil {
		return false
	}
	return k.Available(t)
}

func (b *Balance) Debit(t time.Time, mins int) (int, error) {
	if b.Minutes <= mins {
		n := b.Minutes
		b.Minutes = 0
		return n, ZeroBalance
	}
	b.Minutes -= mins
	return mins, nil
}

type balanceSlice []*Balance

func (s balanceSlice) Len() int           { return len(s) }
func (s balanceSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s balanceSlice) Less(i, j int) bool { return s[i].Priority < s[j].Priority }
