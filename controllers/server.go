package controllers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/symcn/pkg/metrics"
	"github.com/symcn/sym-ops/pkg/debug"
	"k8s.io/klog/v2"
)

func metricServer(metricPort int) *http.Server {
	if metricPort < 1 {
		klog.Warningf("Disabled metrics export")
		return nil
	}

	klog.Info("Enabled metrics export")
	mux := http.NewServeMux()
	metrics.RegisterHTTPHandler(func(pattern string, handler http.Handler) {
		mux.Handle(pattern, handler)
	})
	return &http.Server{
		Addr:    fmt.Sprintf(":%d", metricPort),
		Handler: mux,
	}
}

func pprofServer(pprofPort int) *http.Server {
	if pprofPort < 1 {
		klog.Warningf("Disabled pprof export")
		return nil
	}

	klog.Info("Enabled pprof export")
	mux := http.NewServeMux()
	debug.InitDebug(mux, true)
	return &http.Server{
		Addr:    fmt.Sprintf(":%d", pprofPort),
		Handler: mux,
	}
}

type httpServer struct {
	server *http.Server
}

func (h *httpServer) Start(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		c, cancel := context.WithTimeout(context.TODO(), time.Second*5)
		defer cancel()
		h.server.Shutdown(c)
	}()

	if err := h.server.ListenAndServe(); err != nil {
		if !strings.EqualFold(err.Error(), "http: Server closed") {
			klog.Errorf("start pprof server failed: %+v", err)
		}
	}
	return nil
}
