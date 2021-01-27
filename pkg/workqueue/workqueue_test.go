package workqueue

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/symcn/sym-ops/pkg/metrics"
	"github.com/symcn/sym-ops/pkg/types"
	ktypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
)

type reconcileException struct {
	done    chan struct{}
	count   int
	requeue types.NeedRequeue
	after   time.Duration
	sleep   time.Duration
	err     error
}

func (r *reconcileException) Reconcile(item ktypes.NamespacedName) (types.NeedRequeue, time.Duration, error) {
	klog.Infof("mock Reconcile:%s", item.String())
	if r.sleep > 0 {
		time.Sleep(r.sleep)
	}
	if r.count < 1 {
		close(r.done)
		return types.Done, 0, nil
	}
	r.count--
	return r.requeue, r.after, r.err
}

func TestNewQueueException(t *testing.T) {
	t.Run("return error", func(t *testing.T) {
		done := make(chan struct{}, 0)
		queue, err := NewQueue(&reconcileException{done: done, count: 2, err: errors.New("mock error")}, "return_error", 1, time.Second)
		if err != nil {
			t.Error(err)
			return
		}
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()
		go func() {
			queue.Run(ctx)
		}()

		queue.Add(ktypes.NamespacedName{Namespace: "default", Name: "mock error"})
		<-done
	})

	t.Run("return after", func(t *testing.T) {
		done := make(chan struct{}, 0)
		queue, err := NewQueue(&reconcileException{done: done, count: 5, after: time.Microsecond * 100}, "return_after", 1, time.Second)
		if err != nil {
			t.Error(err)
			return
		}
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()
		go func() {
			queue.Run(ctx)
		}()

		queue.Add(ktypes.NamespacedName{Namespace: "default", Name: "mock error"})
		<-done
	})

	t.Run("return requeue", func(t *testing.T) {
		done := make(chan struct{}, 0)
		queue, err := NewQueue(&reconcileException{done: done, count: 2, requeue: types.Requeue}, "return_requeue", 1, time.Second)
		if err != nil {
			t.Error(err)
			return
		}
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()
		go func() {
			queue.Run(ctx)
		}()

		queue.Add(ktypes.NamespacedName{Namespace: "default", Name: "mock error"})
		<-done
	})

	t.Run("type unexpected", func(t *testing.T) {
		done := make(chan struct{}, 0)
		queue, err := NewQueue(&reconcileException{done: done}, "unexpected_type", 1, time.Second)
		if err != nil {
			t.Error(err)
			return
		}
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()
		go func() {
			queue.Run(ctx)
		}()

		queue.Add("unexpected_type")
		time.Sleep(time.Millisecond * 200)
	})

	t.Run("add after shutdown", func(t *testing.T) {
		done := make(chan struct{}, 0)
		queue, err := NewQueue(&reconcileException{done: done, sleep: time.Millisecond * 100}, "add_after_shutdown", 1, time.Second)
		if err != nil {
			t.Error(err)
			return
		}
		ctx, cancel := context.WithCancel(context.TODO())
		go func() {
			queue.Run(ctx)
		}()
		queue.Add(ktypes.NamespacedName{Namespace: "default", Name: "mock error"})
		cancel()
		queue.Add(ktypes.NamespacedName{Namespace: "default", Name: "mock error"})
	})
}

func TestNewMetrics(t *testing.T) {
	server := startHTTPPrometheus(t)
	defer func() {
		server.Shutdown(context.Background())
	}()

	done := make(chan struct{}, 0)
	count := 10000
	queue, err := NewQueue(&reconcile{done: done, count: count, err: errors.New("mock error")}, "benchmark", 1, time.Second)
	if err != nil {
		t.Error(err)
		return
	}
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	go func() {
		queue.Run(ctx)
	}()

	for i := 0; i < count; i++ {
		queue.Add(ktypes.NamespacedName{Namespace: "default", Name: fmt.Sprintf("item_%d", i)})
		// workqueue_return_requeue_ name_return_requeue_reconcile_fail_total
	}
	<-done
}

type reconcile struct {
	done  chan struct{}
	count int
	sleep time.Duration
	err   error
}

func (r *reconcile) Reconcile(item ktypes.NamespacedName) (types.NeedRequeue, time.Duration, error) {
	klog.Infof("mock Reconcile:%s", item.String())
	r.count--
	if r.count < 1 {
		if r.count == 0 {
			close(r.done)
		}
		return types.Done, 0, nil
	}
	switch r.count % 4 {
	case 0:
		return types.Requeue, 0, nil
	case 1:
		return types.Done, time.Millisecond * 20, nil
	case 2:
		return types.Done, 0, errors.New("mock error")
	case 3:
		time.Sleep(time.Millisecond * 10)
		return types.Done, 0, nil
	}
	return types.Done, 0, nil
}

// startHTTPPrometheus start http server with prometheus route
func startHTTPPrometheus(t *testing.T) *http.Server {
	server := &http.Server{
		Addr: ":8080",
	}
	mux := http.NewServeMux()
	metrics.RegisterHTTPHandler(func(pattern string, handler http.Handler) {
		mux.Handle(pattern, handler)
	})
	server.Handler = mux

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if !strings.EqualFold(err.Error(), "http: Server closed") {
				t.Error(err)
			}
		}
		t.Log("http shutdown")
	}()
	return server
}
