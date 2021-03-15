package ops

import (
	"flag"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"
)

// GetRootCmd returns the root of the cobra command-tree.
func GetRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "sym-ops",
		Short:        "sym-ops",
		Long:         "sym-ops use declarative approach, dvelop project in multiple Kubernetes clusters.",
		SilenceUsage: true,
	}
	// Make sure that klog logging variables are initialized so that we can
	// update them from this file.
	klog.InitFlags(flag.CommandLine)

	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	rootCmd.AddCommand(ControllerCmd())
	rootCmd.AddCommand(VersionCmd())

	return rootCmd
}

// PrintFlags logs the flags in the flagset
func PrintFlags(flags *pflag.FlagSet) {
	flags.VisitAll(func(flag *pflag.Flag) {
		klog.Infof("FLAG: --%s=%q", flag.Name, flag.Value)
	})
}
