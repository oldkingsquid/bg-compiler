package cmd

import (
	"github.com/oldkingsquid/bg-compiler/docker"
	"github.com/oldkingsquid/bg-compiler/flags"
	"github.com/oldkingsquid/bg-compiler/processor"
	"github.com/oldkingsquid/bg-compiler/webserver"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const cmdStartDesc = `Start a compiler webserver that takes HTTP requests and 
runs them in a docker container.`

var cmdStart = &cobra.Command{
	Use:   "start",
	Short: "Start a compiler webserver",
	Long:  cmdStartDesc,
	Args:  cobra.NoArgs,
	Run:   startCmd,
}

func startCmd(cmd *cobra.Command, args []string) {
	docker.Init()
	processor.InitWorkers()

	server := webserver.NewServer(flags.Port())
	logrus.WithError(server.Start()).Fatalln("Error starting server")
}
