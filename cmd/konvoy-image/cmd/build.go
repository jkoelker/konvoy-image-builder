package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/mesosphere/konvoy-image-builder/pkg/app"
	"github.com/mesosphere/konvoy-image-builder/pkg/packer"
)

const workDirFlagName = "work-dir"

type buildCLIFlags struct {
	packerPath         string
	packerManifestPath string
	packerBuildOnError string
	workDir            string
	dryRun             bool

	generateCLIFlags
}

var buildFlags buildCLIFlags

var buildCmd = &cobra.Command{
	Use:   "build <image.yaml>",
	Short: "build and provision images",
	Long: "Build and Provision images. Specifying AWS arguments is deprecated and will " +
		"be removed in a future version. Use the `aws` subcommand instead.",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runBuild(args[0])
	},
}

func newBuilder() *app.Builder {
	return &app.Builder{}
}

func workDir(image string, builder *app.Builder) string {
	if buildFlags.workDir == "" {
		dir, err := builder.InitConfig(newInitOptions(image, buildFlags.generateCLIFlags))
		if err != nil {
			bail("error rendering builder configuration", err, 2)
		}

		return dir
	}

	log.Printf("using workDir provided by --%s flag: %s", workDirFlagName, buildFlags.workDir)
	return buildFlags.workDir
}

func runBuild(image string) {
	builder := newBuilder()
	work := workDir(image, builder)

	if err := builder.Run(work, NewBuildOptions()); err != nil {
		bail("error during run", err, 3)
	}
}

func NewBuildOptions() app.BuildOptions {
	return app.BuildOptions{
		PackerPath: buildFlags.packerPath,
		PackerBuildFlags: packer.BuildFlags{
			Debug:   rootFlags.LogDebug(),
			Color:   rootFlags.Color,
			OnError: buildFlags.packerBuildOnError,
		},
		CustomManifestPath: buildFlags.packerManifestPath,
		DryRun:             buildFlags.dryRun,
	}
}

func init() {
	initBuildAws()

	fs := buildCmd.Flags()

	initGenerateFlags(fs, &buildFlags.generateCLIFlags)
	initAmazonFlags(fs, &buildFlags.generateCLIFlags)

	addBuildArgs(fs, &buildFlags)

	buildCmd.AddCommand(awsBuildCmd)
}

func addBuildArgs(fs *flag.FlagSet, buildArgs *buildCLIFlags) {
	fs.StringVar(
		&buildArgs.packerPath,
		"packer-path",
		packer.DefaultPath,
		"the location of the packer binary",
	)
	fs.StringVar(
		&buildArgs.packerManifestPath,
		"packer-manifest",
		"",
		"provide the path to a custom packer manifest",
	)
	fs.StringVar(
		&buildArgs.packerBuildOnError,
		"packer-on-error",
		"",
		"[advanced] set error strategy for packer. strategies [cleanup, abort, run-cleanup-provisioner]",
	)
	fs.StringVar(
		&buildArgs.workDir,
		workDirFlagName,
		"",
		"path to custom work directory generated by the generate command",
	)
	fs.BoolVar(
		&buildArgs.dryRun,
		"dry-run",
		false,
		"do not create artifacts, or delete them after creating. Recommended for tests.",
	)
}
