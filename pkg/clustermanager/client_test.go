package clustermanager

import (
	"context"
	"testing"
	"time"
)

func TestExceptionNewMingleClient(t *testing.T) {
	t.Run("config is empty", func(t *testing.T) {
		_, err := NewMingleClient(nil)
		if err == nil {
			t.Error("client config is empty, should be error")
		}
	})
	t.Run("cluster configuration is empty", func(t *testing.T) {
		_, err := NewMingleClient(&ClientConfig{})
		if err == nil {
			t.Error("cluster configuration config is empty, should be error")
		}
	})
	t.Run("scheme is empty", func(t *testing.T) {
		cfg := SingleClientConfig(nil)
		cfg.Scheme = nil
		_, err := NewMingleClient(&ClientConfig{})
		if err == nil {
			t.Error("scheme config is empty, should be error")
		}
	})
	t.Run("no health check", func(t *testing.T) {
		cli, err := NewMingleClient(SingleClientConfig(nil))
		if err != nil {
			t.Error(err)
			return
		}
		ctx, cancel := context.WithCancel(context.TODO())
		go func() {
			err = cli.Start(ctx)
		}()

		time.Sleep(time.Millisecond * 100)
		cancel()
		time.Sleep(time.Millisecond * 100)

		if err != nil {
			t.Error(err)
		}
	})
}

func TestNewMingleClient(t *testing.T) {

}
