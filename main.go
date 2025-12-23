package main

import (
	"github.com/oldkingsquid/bg-compiler/cmd"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := cmd.Execute(); err != nil {
		logrus.WithError(err).Fatalf("Error executing flags")
	}
}
