package main

import (
	"github.com/pefish/go-commander"
	go_logger "github.com/pefish/go-logger"
	"github.com/pefish/nucypher-node/cmd/nucypher-node/command"
	"github.com/pefish/nucypher-node/version"
)

func main() {
	commanderInstance := commander.NewCommander(version.AppName, version.Version, version.AppName+" 是一个 nucypher 挖矿节点，祝你玩得开心。作者：pefish")
	commanderInstance.RegisterDefaultSubcommand(command.NewDefaultCommand())
	err := commanderInstance.Run()
	if err != nil {
		go_logger.Logger.Error(err)
	}
}
