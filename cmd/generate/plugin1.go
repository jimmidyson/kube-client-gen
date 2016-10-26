package generate

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jimmidyson/kube-client-gen/pkg/log"
)

var plugin1 = &cobra.Command{
	Use:   "plugin1",
	Short: "Plugin1",
	Run: func(cmd *cobra.Command, args []string) {
		log.Log.Debug("generating for", "packages", fmt.Sprintf("%#v", parsedPackages))
	},
}

func init() {
	RootCmd.AddCommand(plugin1)
}
