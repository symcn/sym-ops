package ops

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/symcn/sym-ops/pkg/version"
)

var (
	format string = `
   _______  ______ ___        ____  ____  _____
  / ___/ / / / __ ` + "`" + `__ \______/ __ \/ __ \/ ___/
 (__  ) /_/ / / / / / /_____/ /_/ / /_/ (__  ) 
/____/\__, /_/ /_/ /_/      \____/ .___/____/  
     /____/                     /_/

Release:      %s
Commit:       %s
Date:         %s
GoVersion:    %s
Platform:     %s
`
)

// VersionCmd version component
func VersionCmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:  "version",
		Long: "Print version/build info",
		RunE: func(cmd *cobra.Command, args []string) error {
			v := version.GetVersion()
			fmt.Fprintf(os.Stdout, format, v.Release, v.GitCommit, v.BuildDate, v.GoVersion, v.Platform)
			return nil
		},
	}
	return versionCmd
}
