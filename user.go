package main

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"time"
)

const deadTimeout = time.Minute * 1

var ZeroBalance = errors.New("no minutes available")

type User struct {
	Balance  []*Balance
	LastSeen time.Time
}

func (u *User) Running(t time.Time) bool {
	return t.Sub(u.LastSeen) < deadTimeout
}

func (u *User) CheckIn(t time.Time) bool {
	u.LastSeen = t
	return u.AvailableBalance(t) != nil
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

func (u *User) SetBalance(kind string, minutes int) {
	for _, b := range u.Balance {
		if b.Kind != kind {
			continue
		}
		b.Minutes = minutes
	}
}

type Balance struct {
	Kind     string
	Minutes  int
	Priority int // balances with lower priorities will be drained first
}

func (b *Balance) Available(t time.Time) bool {
	k := kindMap[b.Kind]
	if k == nil {
		log.Println("unknown kind:", b.Kind)
		return false
	}
	return k.Available(t) && b.Minutes > 0
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

func (b *Balance) String() string {
	k, ok := kindMap[b.Kind]
	if !ok {
		panic("unknown kind " + b.Kind)
	}
	return fmt.Sprintf("%s (priority %d): %d minutes (%s)", b.Kind, b.Priority, b.Minutes, k.Times())
}

type balanceSlice []*Balance

func (s balanceSlice) Len() int           { return len(s) }
func (s balanceSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s balanceSlice) Less(i, j int) bool { return s[i].Priority < s[j].Priority }
