package ops

import (
	"github.com/spf13/cobra"
	"github.com/symcn/sym-ops/controllers"
	"github.com/symcn/sym-ops/pkg/types"
)

// ControllerCmd controller component
func ControllerCmd() *cobra.Command {
	opt := controllers.DefaultOptions()
	controllerCmd := &cobra.Command{
		Use:   "controller",
		Short: "Start controller component.",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			PrintFlags(cmd.Flags())

			ctrl, err := controllers.NewControllers(opt)
			if err != nil {
				return err
			}
			return ctrl.Start()
		},
	}

	// Controller config
	controllerCmd.PersistentFlags().IntVar(&opt.Threadiness, "threadiness", opt.Threadiness, "the max goroutine for Reconcile")
	controllerCmd.PersistentFlags().IntVar(&opt.MetricPort, "metric-port", opt.MetricPort, "metric listener port, 0 means close metric")
	controllerCmd.PersistentFlags().IntVar(&opt.PprofPort, "pprof-port", opt.PprofPort, "pprof listener port, 0 means close pprof")
	controllerCmd.PersistentFlags().BoolVar(&opt.Master, "master", opt.Master, "enable master feature")
	controllerCmd.PersistentFlags().BoolVar(&opt.Worker, "worker", opt.Worker, "enable worker feature")

	// ClusterManagerOptions config
	controllerCmd.PersistentFlags().IntVar(&opt.ClusterManagerOptions.QPS, "qps", opt.ClusterManagerOptions.QPS, "maximum QPS to the master from this client")
	controllerCmd.PersistentFlags().IntVar(&opt.ClusterManagerOptions.Burst, "burst", opt.ClusterManagerOptions.Burst, "maximum burst for throttle")
	controllerCmd.PersistentFlags().BoolVar(&opt.ClusterManagerOptions.LeaderElection, "leader", opt.ClusterManagerOptions.LeaderElection, "enable leader election")
	controllerCmd.PersistentFlags().StringVar(&opt.ClusterManagerOptions.LeaderElectionNamespace, "leader-ns", opt.ClusterManagerOptions.LeaderElectionNamespace, "leader election with namespace")
	controllerCmd.PersistentFlags().StringVar(&opt.ClusterManagerOptions.LeaderElectionID, "leader-id", opt.ClusterManagerOptions.LeaderElectionID, "leader election with id")
	controllerCmd.PersistentFlags().DurationVar(&opt.ClusterManagerOptions.SyncPeriod, "sync-period", opt.ClusterManagerOptions.SyncPeriod, "sync period determines the minimum frequency at which watched resource are reconciled")
	controllerCmd.PersistentFlags().DurationVar(&opt.ClusterManagerOptions.HealthCheckInterval, "health-interval", opt.ClusterManagerOptions.HealthCheckInterval, "Kubernetes connected health check interval, 0 means close health check")
	controllerCmd.PersistentFlags().DurationVar(&opt.ClusterManagerOptions.ExecTimeout, "exec-timeout", opt.ClusterManagerOptions.ExecTimeout, "exec with timeout")

	// Advdeployment config
	controllerCmd.PersistentFlags().Int32Var(&opt.AdvConfig.RevisionHistoryLimit, "revision-limit", opt.AdvConfig.RevisionHistoryLimit, "revision history limit")
	controllerCmd.PersistentFlags().Int32Var(&opt.AdvConfig.ProgressDeadlineSeconds, "progress-deadline-seconds", opt.AdvConfig.ProgressDeadlineSeconds, "progress-deadline-seconds")

	// namespace filter
	controllerCmd.PersistentFlags().StringArrayVar(&types.FilterNamespaceAppset, "filter-namespace-master", types.FilterNamespaceAppset, "master watch resource filter namespace")
	controllerCmd.PersistentFlags().StringArrayVar(&types.FilterNamespaceAdvdeployment, "filter-namespace-worker", types.FilterNamespaceAdvdeployment, "worker watch resource filter namespace")

	return controllerCmd
}
