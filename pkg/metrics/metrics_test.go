package metrics

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestNewMetrics(t *testing.T) {
	defer resetAll()

	metrics, err := NewMetrics("sym-ops", map[string]string{
		"k1":  "v1",
		"k2":  "v2",
		"k3":  "v3",
		"k4":  "v4",
		"k5":  "v5",
		"k6":  "v6",
		"k7":  "v7",
		"k8":  "v8",
		"k9":  "v9",
		"k10": "v10",
		"k11": "v11",
		"k12": "v12",
		"k13": "v13",
		"k14": "v14",
		"k15": "v15",
		"k16": "v16",
		"k17": "v17",
		"k18": "v18",
		"k19": "v19",
		"k20": "v20",
		"k21": "v21",
	})
	if err == nil {
		t.Error("max label should exception")
		return
	}

	metrics, err = NewMetrics("sym-ops", map[string]string{
		"k1": "v1",
	})
	if err != nil {
		t.Error(err)
		return
	}

	metrics.Counter("counter").Add(1)
	metrics.Gauge("gauge").Add(1)
	metrics.Histogram("histogram", nil).Observe(1)
	metrics.Histogram("summary", nil).Observe(1)
	metrics.Counter("repeat").Add(1)
	metrics.Counter("repeat").Add(1)
}

func TestRegisterHTTPRoute(t *testing.T) {
	defer resetAll()

	resetAll()

	metrics, err := NewMetrics("sym-ops", map[string]string{
		"k1": "v1",
	})
	if err != nil {
		t.Error(err)
		return
	}

	server := startHTTPPrometheus(t)
	defer func() {
		server.Shutdown(context.Background())
	}()

	interval := time.Millisecond * 100
	metrics.Counter("counter").Add(1)
	time.Sleep(interval)
	body, err := request()
	if err != nil {
		t.Error(err)
		return
	}
	if !strings.Contains(body, "sym_ops_k1_v1_counter") {
		t.Error("counter not register")
		return
	}
	if strings.Contains(body, "sym_ops_k1_v1_gauge") {
		t.Error("gauge not register, shouldn't exist")
		return
	}

	metrics.Gauge("gauge").Add(1)
	time.Sleep(interval)
	body, err = request()
	if err != nil {
		t.Error(err)
		return
	}
	if !strings.Contains(body, "sym_ops_k1_v1_gauge") {
		t.Error("gauge not register")
		return
	}

	metrics.Histogram("histogram", nil).Observe(1)
	time.Sleep(interval)
	body, err = request()
	if err != nil {
		t.Error(err)
		return
	}
	if !strings.Contains(body, "sym_ops_k1_v1_histogram") {
		t.Error("histogram not register")
		return
	}
}

// startHTTPPrometheus start http server with prometheus route
func startHTTPPrometheus(t *testing.T) *http.Server {
	server := &http.Server{
		Addr: ":8080",
	}
	mux := http.NewServeMux()
	RegisterHTTPHandler(func(pattern string, handler http.Handler) {
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

func request() (data string, err error) {
	resp, err := http.Get("http://localhost:8080" + defaultEndpoint)
	if err != nil || resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("request fail")
		return "", err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return string(body), nil
}
