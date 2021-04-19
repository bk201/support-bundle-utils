package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/bk201/support-bundle-utils/pkg/manager"
	"github.com/bk201/support-bundle-utils/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	sbm = &manager.SupportBundleManager{}
)

// managerCmd represents the manager command
var managerCmd = &cobra.Command{
	Use:   "manager",
	Short: "Harvester support bundle manager.",
	Long: `Harvester support bundle manager.

The manager collects following items:
- Cluster level bundle. Including resource manifests and pod logs.
- Any external bundles. e.g., Longhorn support bundle.

And it also waits for reports from support bundle agents. The reports contain:
- Logs of each Harvester node.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := sbm.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(managerCmd)
	managerCmd.PersistentFlags().StringVar(&sbm.Namespace, "namespace", os.Getenv("HARVESTER_NAMESPACE"), "The Harvester namespace")
	managerCmd.PersistentFlags().StringVar(&sbm.BundleName, "bundlename", os.Getenv("HARVESTER_SUPPORT_BUNDLE_NAME"), "The support bundle name")
	managerCmd.PersistentFlags().StringVar(&sbm.OutputDir, "outdir", os.Getenv("HARVESTER_SUPPORT_BUNDLE_OUTPUT_DIR"), "The directory to store the bundle")

	nodeCount := utils.EnvGetInt("HARVESTER_SUPPORT_BUNDLE_NODE_COUNT", 0)
	managerCmd.PersistentFlags().IntVar(&sbm.NodeCount, "nodecount", nodeCount, "The number of node bundles to wait")

	timeout := utils.EnvGetDuration("HARVESTER_SUPPORT_BUNDLE_WAIT_TIMEOUT", 5*time.Minute)
	managerCmd.PersistentFlags().DurationVar(&sbm.WaitTimeout, "wait", timeout, "The timeout to wait for node bundles")

	managerCmd.PersistentFlags().StringVar(&sbm.LonghornAPI, "longhorn-api", "http://longhorn-backend.longhorn-system:9500", "The Longhorn API URL")
}
