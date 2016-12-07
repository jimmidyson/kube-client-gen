package generate

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jimmidyson/kube-client-gen/pkg/generator/immutables"
)

var (
	immutablesCmd = &cobra.Command{
		Use:   "immutables",
		Short: "Java Immutables",
		Run: func(cmd *cobra.Command, args []string) {
			immConfig := immutables.Config{
				Config:                   config,
				JavaRootPackage:          *javaRootPackage,
				StyleClass:               *stylesClass,
				JavaRootOpenShiftPackage: *javaRootOpenShiftPackage,
			}
			gen := immutables.New(immConfig)
			err := gen.Generate(parsedPackages)
			if err != nil {
				config.Logger.Crit("failed to generate", "type", "immutables", "error", err)
				os.Exit(1)
			}
		},
	}

	defaultJavaRootPackage = "io.fabric8.kubernetes.types"
	javaRootPackage        *string

	defaultJavaRootOpenShiftPackage = "io.fabric8.openshift.types"
	javaRootOpenShiftPackage        *string

	defaultStylesClass = strings.Join([]string{defaultJavaRootPackage, "common", "ImmutablesStyle"}, ".")
	stylesClass        *string
)

func init() {
	javaRootPackage = immutablesCmd.Flags().StringP("java-root-package", "j", defaultJavaRootPackage, "root java package to generate Kubernetes classes in")
	javaRootOpenShiftPackage = immutablesCmd.Flags().String("java-root-openshift-package", defaultJavaRootOpenShiftPackage, "root java package to generate OpenShift classes in")
	stylesClass = immutablesCmd.Flags().StringP("styles-class", "s", defaultStylesClass, "default immutables styles class")

	RootCmd.AddCommand(immutablesCmd)
}
