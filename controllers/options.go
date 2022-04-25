package controllers

import (
	"time"

	"github.com/symcn/api"
	"github.com/symcn/pkg/clustermanager/client"
	workloadv1beta1 "github.com/symcn/sym-ops/api/v1beta1"
	"github.com/symcn/sym-ops/controllers/advdeployment"
	"github.com/symcn/sym-ops/pkg/types"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

func init() {
	clientgoscheme.AddToScheme(types.Scheme)
	apiextensionsv1beta1.AddToScheme(types.Scheme)
	workloadv1beta1.AddToScheme(types.Scheme)
}

// Options controllers options
type Options struct {
	ClusterManagerOptions *client.Options

	Threadiness int
	GotInterval time.Duration
	MetricPort  int
	PprofPort   int

	Master bool
	Worker bool

	AdvConfig *advdeployment.AdvConfig
}

// DefaultOptions default controllers options
func DefaultOptions() *Options {
	opt := client.DefaultOptionsWithScheme(types.Scheme)
	opt.SetKubeRestConfigFnList = []api.SetKubeRestConfig{
		func(cfg *rest.Config) {
			cfg.UserAgent = "sym-ops-controller"
		},
	}

	return &Options{
		ClusterManagerOptions: opt,
		Threadiness:           1,
		GotInterval:           time.Second * 1,
		MetricPort:            9090,
		PprofPort:             34901,
		Master:                false,
		Worker:                false,
		AdvConfig:             advdeployment.DefaultAdvConfig(),
	}
}
