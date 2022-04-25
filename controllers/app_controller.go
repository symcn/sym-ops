package controllers

import (
	"net/http"

	"github.com/symcn/api"
	"github.com/symcn/pkg/clustermanager/client"
	"github.com/symcn/pkg/metrics"
	"github.com/symcn/sym-ops/controllers/advdeployment"
	"github.com/symcn/sym-ops/controllers/appset"
	"github.com/symcn/sym-ops/pkg/debug"
	"github.com/symcn/sym-ops/pkg/types"
	"github.com/symcn/sym-ops/pkg/utils"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

// App controllers
type App struct {
	*Options

	currentCli api.MingleClient
	server     *utils.Server
}

// NewControllers build App
func NewControllers(opt *Options) (*App, error) {
	app := &App{
		Options: opt,
		server:  &utils.Server{},
	}
	currentCli, err := client.NewMingleClient(client.DefaultClusterCfgInfo(types.CurrentClusterName), app.ClusterManagerOptions)
	if err != nil {
		return nil, err
	}
	app.server.Add(currentCli)

	// add metrics service
	if srv := metricServer(opt.MetricPort); srv != nil {
		app.server.Add(&httpServer{server: srv})
	}
	// add pprof service
	if srv := pprofServer(opt.PprofPort); srv != nil {
		app.server.Add(&httpServer{server: srv})
	}

	if opt.Master {
		err = appset.MasterFeature(currentCli, app.Threadiness, app.GotInterval, app.server, app.ClusterManagerOptions)
		if err != nil {
			return nil, err
		}
	}
	if opt.Worker {
		err = advdeployment.WorkerFeature(currentCli, app.Threadiness, app.GotInterval, app.AdvConfig, app.server, app.ClusterManagerOptions)
		if err != nil {
			return nil, err
		}
	}

	return app, nil
}

// Start start controller
func (app *App) Start() error {
	// start http server

	mux := http.NewServeMux()
	metrics.RegisterHTTPHandler(func(pattern string, handler http.Handler) {
		mux.Handle(pattern, handler)
	})
	debug.InitDebug(mux, true)

	ctx := signals.SetupSignalHandler()
	if err := app.server.Start(ctx); err != nil {
		klog.Error(err)
		return err
	}
	return nil
}
