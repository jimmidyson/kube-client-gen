package generate

import (
	"os"

	"github.com/inconshreveable/log15"
	"github.com/spf13/cobra"

	"github.com/jimmidyson/kube-client-gen/pkg/generator"
	"github.com/jimmidyson/kube-client-gen/pkg/loader"
	"github.com/jimmidyson/kube-client-gen/pkg/log"
)

var (
	defaultAPIPackages = []string{
		"k8s.io/kubernetes/pkg/api/unversioned",
		"k8s.io/kubernetes/pkg/api/resource",

		"k8s.io/kubernetes/pkg/api/v1",

		"k8s.io/kubernetes/pkg/apis/apps/v1alpha1",
		"k8s.io/kubernetes/pkg/apis/authentication/v1beta1",
		"k8s.io/kubernetes/pkg/apis/autoscaling/v1",
		"k8s.io/kubernetes/pkg/apis/batch/v1",
		"k8s.io/kubernetes/pkg/apis/batch/v2alpha1",
		"k8s.io/kubernetes/pkg/apis/extensions/v1beta1",
		"k8s.io/kubernetes/pkg/apis/policy/v1alpha1",
		"k8s.io/kubernetes/pkg/apis/rbac/v1alpha1",

		"github.com/openshift/origin/pkg/authorization/api/v1",
		"github.com/openshift/origin/pkg/build/api/v1",
		"github.com/openshift/origin/pkg/deploy/api/v1",
		"github.com/openshift/origin/pkg/image/api/v1",
		"github.com/openshift/origin/pkg/oauth/api/v1",
		"github.com/openshift/origin/pkg/project/api/v1",
		"github.com/openshift/origin/pkg/quota/api/v1",
		"github.com/openshift/origin/pkg/route/api/v1",
		"github.com/openshift/origin/pkg/sdn/api/v1",
		"github.com/openshift/origin/pkg/security/api/v1",
		"github.com/openshift/origin/pkg/template/api/v1",
		"github.com/openshift/origin/pkg/user/api/v1",
	}

	RootCmd = &cobra.Command{
		Use:   "kube-client-gen",
		Short: "Kubernetes Client Generator",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logger := log.Log

			logger.SetHandler(log15.CallerFileHandler(log15.StderrHandler))

			logLvl := defaultLogLevel
			if *verbose {
				logLvl = log15.LvlDebug
			}
			logger.SetHandler(log15.LvlFilterHandler(logLvl, log.Log.GetHandler()))

			config = generator.Config{
				Logger:          logger,
				Force:           *force,
				OutputDirectory: *outputDirectory,
			}

			ldr := loader.New(*packages, logger)
			pkgs, err := ldr.Load()
			if err != nil {
				logger.Error("failed to parse packages", "error", err)
				os.Exit(1)
			}
			parsedPackages = pkgs
		},
	}

	packages        *[]string
	verbose         *bool
	outputDirectory *string
	force           *bool

	defaultLogLevel = log15.LvlInfo
	config          generator.Config
	parsedPackages  []loader.Package
)

const pluginBinaryPrefix = "kmg-"

func init() {
	packages = RootCmd.PersistentFlags().StringSliceP("package", "p", defaultAPIPackages, "packages to generate JSON schema for")
	verbose = RootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	outputDirectory = RootCmd.Flags().StringP("output-directory", "o", "", "the directory to output generated files to")
	force = RootCmd.Flags().BoolP("force", "f", false, "force overwrite of existing files")
}
