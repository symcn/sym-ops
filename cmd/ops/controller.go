package ops

import (
	"context"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	workloadv1beta1 "github.com/symcn/sym-ops/api/v1beta1"
	"github.com/symcn/sym-ops/controllers/appset"
	"github.com/symcn/sym-ops/pkg/clustermanager"
	"github.com/symcn/sym-ops/pkg/metrics"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	clientgoscheme.AddToScheme(scheme)
	apiextensionsv1beta1.AddToScheme(scheme)
	workloadv1beta1.AddToScheme(scheme)
}

// ControllerCmd controller component
func ControllerCmd() *cobra.Command {
	opt := defaultCtrlOption()

	controllerCmd := &cobra.Command{
		Use:   "controller",
		Short: "Start controller component.",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			PrintFlags(cmd.Flags())

			metaCli, err := clustermanager.NewMingleClient(clustermanager.DefaultClusterCfgInfo("meta"), clustermanager.DefaultOptions(scheme, opt.Qos, opt.Burst))
			if err != nil {
				return err
			}
			appset.MasterFeature(metaCli)

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
						klog.Error(err)
					}
				}
			}()

			ctx := context.TODO()
			if err = metaCli.Start(ctx); err != nil {
				klog.Error(err)
			}
			return nil
		},
	}

	controllerCmd.PersistentFlags().IntVar(&opt.Threadiness, "threadiness", opt.Threadiness, "the max goroutine for Reconcile")
	controllerCmd.PersistentFlags().IntVar(&opt.MetricPort, "metric-port", opt.MetricPort, "metric listener port, 0 means close metric")
	controllerCmd.PersistentFlags().IntVar(&opt.PprofPort, "pprof-port", opt.PprofPort, "pprof listener port, 0 means close pprof")
	controllerCmd.PersistentFlags().IntVar(&opt.Qos, "qos", opt.Qos, "maximum QPS to the master from this client")
	controllerCmd.PersistentFlags().IntVar(&opt.Burst, "burst", opt.Burst, "maximum burst for throttle")

	controllerCmd.PersistentFlags().BoolVar(&opt.LeaderElection, "leader", opt.LeaderElection, "enable leader election")
	controllerCmd.PersistentFlags().StringVar(&opt.LeaderElectionNamespace, "leader-ns", opt.LeaderElectionNamespace, "leader election with namespace")
	controllerCmd.PersistentFlags().StringVar(&opt.LeaderElectionID, "leader-id", opt.LeaderElectionID, "leader election with id")

	controllerCmd.PersistentFlags().BoolVar(&opt.Master, "master", opt.Master, "enable master feature")
	controllerCmd.PersistentFlags().BoolVar(&opt.Worker, "worker", opt.Worker, "enable worker feature")

	controllerCmd.PersistentFlags().DurationVar(&opt.SyncPeriod, "sync-period", opt.SyncPeriod, "sync period determines the minimum frequency at which watched resource are reconciled")
	controllerCmd.PersistentFlags().DurationVar(&opt.HealthCheckInterval, "health-interval", opt.HealthCheckInterval, "Kubernetes connected health check interval, 0 means close health check")
	controllerCmd.PersistentFlags().DurationVar(&opt.ExecTimeout, "exec-timeout", opt.ExecTimeout, "exec with timeout")

	return controllerCmd
}
