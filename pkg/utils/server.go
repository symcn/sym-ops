package utils

import (
	"context"
	"sync"

	"k8s.io/klog/v2"
)

// Runnable define server
type Runnable interface {
	Start(ctx context.Context) error
}

// startFunc defines a function that will be used to start one or more components of the service.
type startFunc func(ctx context.Context) error

// Server server list
type Server struct {
	l       sync.Mutex
	servers []startFunc
	wg      sync.WaitGroup
}

// Add add Runnable service
func (s *Server) Add(r Runnable) {
	s.l.Lock()
	defer s.l.Unlock()

	if len(s.servers) == 0 {
		s.servers = []startFunc{}
	}
	s.servers = append(s.servers, func(ctx context.Context) error {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()

			if err := r.Start(ctx); err != nil {
				klog.Errorf("failure in startup function: %v", err)
			}
		}()
		return nil
	})
	return
}

// Start start service
func (s *Server) Start(ctx context.Context) error {
	for _, fn := range s.servers {
		if err := fn(ctx); err != nil {
			return err
		}
	}
	s.wg.Wait()
	return nil
}
