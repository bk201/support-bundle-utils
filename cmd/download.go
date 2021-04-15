package cmd

import (
	"fmt"
	"os"

	"github.com/bk201/support-bundle-utils/pkg/client"
	"github.com/spf13/cobra"
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download [api_url]",
	Short: "Generate and download a support bundle from a Harvester cluster",
	Long:  "Generate and download a support bundle from a Harvester cluster",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmdConfig.Run(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "fail to download support bundle: %s\n", err)
			os.Exit(1)
		}
	},
	Args: cobra.ExactArgs(1),
}

var cmdConfig = client.SupportBundleClient{}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.PersistentFlags().BoolVar(&cmdConfig.NoAuth, "noauth", false, "authorize before getting the bundle")
	downloadCmd.PersistentFlags().StringVar(&cmdConfig.User, "user", "", "username")
	downloadCmd.PersistentFlags().StringVar(&cmdConfig.Password, "password", "", "password")
	downloadCmd.PersistentFlags().StringVar(&cmdConfig.OutputFile, "output", "", "output file path (default ${bundle_name}.zip)")
	downloadCmd.PersistentFlags().BoolVar(&cmdConfig.Insecure, "insecure", false, "do not verify server certificate")
	downloadCmd.PersistentFlags().StringVar(&cmdConfig.IssueURL, "issue", "", "issue URL")
	downloadCmd.PersistentFlags().StringVar(&cmdConfig.IssueDescription, "description", "No description", "issue description")
}
