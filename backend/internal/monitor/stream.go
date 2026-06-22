package monitor

import (
	"context"
	"sync"
	"time"
)

type Streamer struct {
	clients map[chan SystemStatus]struct{}
	mutex   sync.RWMutex
}

func NewStreamer() *Streamer {
	return &Streamer{
		clients: make(map[chan SystemStatus]struct{}),
	}
}

func (s *Streamer) Subscribe() chan SystemStatus {
	ch := make(chan SystemStatus, 10)
	s.mutex.Lock()
	s.clients[ch] = struct{}{}
	s.mutex.Unlock()
	return ch
}

func (s *Streamer) Unsubscribe(ch chan SystemStatus) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.clients[ch]; exists {
		delete(s.clients, ch)
		close(ch)
	}
}

func (s *Streamer) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.fetchAndBroadcast(ctx)
		case <-ctx.Done():
			s.cleanup()
			return
		}
	}
}

func (s *Streamer) fetchAndBroadcast(ctx context.Context) {
	status, err := GetSystemStatus(ctx)
	if err != nil {
		return
	}
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	for ch := range s.clients {
		select {
		case ch <- status:
		default:
			// Drop message if channel is full to prevent blocking
		}
	}
}

func (s *Streamer) cleanup() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for ch := range s.clients {
		close(ch)
	}
	s.clients = make(map[chan SystemStatus]struct{})
}
