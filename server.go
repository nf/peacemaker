package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sync"
	"time"
)

type Server struct {
	User   map[string]*User
	Refill []*Refill

	datafile string
	mu       sync.Mutex
}

func NewServer(datafile string) (*Server, error) {
	s := &Server{
		User:     make(map[string]*User),
		datafile: datafile,
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	go s.loop()
	return s, nil
}

func (s *Server) load() error {
	f, err := os.Open(s.datafile)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(s)
}

func (s *Server) save() error {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.OpenFile(s.datafile, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(b)
	return err
}

func (s *Server) loop() {
	tick := time.NewTicker(time.Minute)
	for t := range tick.C {
		if s.tick(t) {
			if err := s.save(); err != nil {
				log.Fatalf("saving: %v", err)
			}
		}
	}
}

func (s *Server) tick(t time.Time) (acted bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, r := range s.Refill {
		if r.Do(t, s.User) {
			acted = true
		}
	}
	for name, u := range s.User {
		if !u.Running(t) {
			continue
		}
		// Debit user 1 minute.
		if err := u.Debit(t, 1); err != nil {
			log.Printf("stopping %s: %v", name, err)
		}
		acted = true
	}
	return
}

func (s *Server) CheckIn(username *string, ok *bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	u := s.User[*username]
	if u == nil {
		return errors.New("user not found")
	}
	*ok = u.CheckIn(time.Now())
	return nil
}

func (s *Server) setBalance(username, kind string, minutes int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	u := s.User[username]
	if u == nil {
		return errors.New("user not found")
	}
	u.SetBalance(kind, minutes)
	return nil
}
