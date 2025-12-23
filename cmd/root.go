package cmd

import (
	"github.com/oldkingsquid/bg-compiler/flags"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var verbose bool

var rootCmd = &cobra.Command{
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verbose {
			logrus.SetLevel(logrus.DebugLevel)
		}
	},
	PreRun: func(cmd *cobra.Command, args []string) {},
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Fatalf("Invalid command")
	},
	PostRun:           func(cmd *cobra.Command, args []string) {},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")

	// Start Command Flags
	cmdStart.Flags().IntVar(&flags.FlagConfig.WorkerCount,
		"workers",
		flags.FlagConfig.WorkerCount,
		"Number of workers to run")
	cmdStart.Flags().IntVar(&flags.FlagConfig.ContainerTimeoutSeconds,
		"timeout",
		flags.FlagConfig.ContainerTimeoutSeconds,
		"Container timeout in seconds")
	cmdStart.Flags().IntVar(&flags.FlagConfig.JobChannelLength,
		"backlog",
		flags.FlagConfig.JobChannelLength,
		"Job backlog before rejecting requests")
	cmdStart.Flags().Int64Var(&flags.FlagConfig.ContainerMaxMemoryMB,
		"memory",
		flags.FlagConfig.ContainerMaxMemoryMB,
		"Maximum amount of memory that a container can use for a single run (in MB)")
	cmdStart.Flags().Int64Var(&flags.FlagConfig.ContainerCPUShares,
		"cpu",
		flags.FlagConfig.ContainerCPUShares,
		"CPU shares for a container")
	cmdStart.Flags().IntVar(&flags.FlagConfig.MaxReadOutputBytesKB,
		"output",
		flags.FlagConfig.MaxReadOutputBytesKB,
		"Maximum number of bytes that can be read from a container output before the container is killed (in KB)")
	cmdStart.Flags().IntVar(&flags.FlagConfig.Port,
		"port",
		flags.FlagConfig.Port,
		"Server port to listen on")
	cmdStart.Flags().BoolVar(&flags.FlagConfig.UseGVisor,
		"gvisor",
		flags.FlagConfig.UseGVisor,
		"Use gVisor for increased container isolation (requires runsc).")

	rootCmd.AddCommand(cmdStart)
}

func Execute() error {
	return rootCmd.Execute()
}
