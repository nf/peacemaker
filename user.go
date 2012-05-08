package main

import (
	"errors"
	"sort"
	"time"
)

var ZeroBalance = errors.New("no minutes available")

type User struct {
	Balance   []*Balance
	StartTime time.Time
	LastSeen  time.Time
}

func (u *User) Start() error {
	t := time.Now()
	if err := u.Debit(t, 1); err != nil {
		return err
	}
	u.StartTime = t
	u.LastSeen = t
	return nil
}

func (u *User) Running() bool {
	return !u.StartTime.IsZero()
}

func (u *User) Stop() {
	u.StartTime = time.Time{}
}

func (u *User) Debit(t time.Time, mins int) error {
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
		if b.Available(t) && !b.Expired(t) {
			return b
		}
	}
	return nil
}

type Balance struct {
	Kind     string
	Minutes  int
	Expiry   time.Time
	Priority int // balances with lower priorities will be drained first
}

func (b *Balance) Available(t time.Time) bool {
	k := kindMap[b.Kind]
	if k == nil {
		return false
	}
	return k.Available(t)
}

func (b *Balance) Expired(t time.Time) bool {
	if b.Expiry.IsZero() {
		return false // no expiry
	}
	return t.After(b.Expiry)
}

func (b *Balance) Debit(t time.Time, mins int) (int, error) {
	if !b.Available(t) {
		return 0, errors.New("balance not available at this time")
	}
	if b.Expired(t) {
		return 0, errors.New("balance expired")
	}
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
