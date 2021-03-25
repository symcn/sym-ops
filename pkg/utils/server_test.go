package utils

import (
	"context"
	"errors"
	"testing"
)

var (
	mockErrorMsg = "mock error"
)

type run struct{}
type runerr struct{}

func (r *run) Start(ctx context.Context) error {
	return nil
}

func (r *runerr) Start(ctx context.Context) error {
	return errors.New(mockErrorMsg)
}

func TestServer(t *testing.T) {
	s1 := &Server{}
	s1.Add(&run{})
	s1.Add(&runerr{})
	err := s1.Start(context.TODO())
	if err != nil {
		t.Errorf("start Runnable have error: %+v", err)
		return
	}
}
