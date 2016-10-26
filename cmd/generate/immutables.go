package generate

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/jimmidyson/kube-client-gen/pkg/generator/immutables"
)

var (
	immutablesCmd = &cobra.Command{
		Use:   "immutables",
		Short: "Java Immutables",
		Run: func(cmd *cobra.Command, args []string) {
			immConfig := immutables.Config{
				Config:          config,
				JavaRootPackage: *javaRootPackage,
			}
			gen := immutables.New(immConfig)
			err := gen.Generate(parsedPackages)
			if err != nil {
				config.Logger.Crit("failed to generate", "type", "immutables", "error", err)
				os.Exit(1)
			}
		},
	}

	defaultJavaRootPackage = "io.fabric8"
	javaRootPackage        *string
)

func init() {
	javaRootPackage = immutablesCmd.Flags().StringP("java-root-package", "j", defaultJavaRootPackage, "root java package to generate classes in")

	RootCmd.AddCommand(immutablesCmd)
}
