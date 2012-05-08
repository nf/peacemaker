package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sync"
	"time"
)

const (
	gracePeriod = time.Minute * 1
	deadTimeout = time.Minute * 1
)

type Server struct {
	User     map[string]*User
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
	return json.NewDecoder(f).Decode(&s.User)
}

func (s *Server) save() error {
	f, err := os.OpenFile(s.datafile, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(&s.User)
}

func (s *Server) loop() {
	t := time.NewTicker(time.Minute)
	for _ = range t.C {
		if s.tick() {
			if err := s.save(); err != nil {
				log.Fatalf("saving: %v")
			}
		}
	}
}

func (s *Server) tick() (acted bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for name, u := range s.User {
		if !u.Running() {
			continue
		}
		// Don't deduct minutes during grace period.
		if now.Sub(u.StartTime) < gracePeriod {
			continue
		}

		acted = true

		// Stop session if user hasn't checked in recently.
		if now.Sub(u.LastSeen) > deadTimeout {
			log.Printf("stopping %s for inactivity")
			u.Stop()
			continue
		}

		// Debit user 1 minute.
		if err := u.Debit(now, 1); err != nil {
			log.Printf("debiting %s: %v", name, err)
			u.Stop()
		}
	}
	return
}

func (s *Server) Start(username *string, _ *struct{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	u := s.User[*username]
	if u == nil {
		return errors.New("user not found")
	}
	return u.Start()
}

func (s *Server) Stop(username *string, _ *struct{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	u := s.User[*username]
	if u == nil {
		return errors.New("user not found")
	}
	u.Stop()
	return nil
}

func (s *Server) IsOpen(username *string, ok *bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	u := s.User[*username]
	if u == nil {
		return errors.New("user not found")
	}
	*ok = u.Running()
	return nil
}

func (s *Server) CheckIn(username *string, ok *bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	u := s.User[*username]
	if u == nil {
		return errors.New("user not found")
	}
	*ok = u.Running()
	u.LastSeen = time.Now()
	return nil
}
