package main

import (
	"github.com/crain-cn/event-mesh/cmd/module"
	"github.com/crain-cn/event-mesh/pkg/logging"
	"os"
)

func main() {
	logging.InitLogger()
	options := module.ParseOptions()
	config := module.ParseConfigYaml()
	memProvider, marker := module.SetALertMemProvider()
	module.SetupK8s(options, config, memProvider)
	os.Exit(module.RunAlertDispatch(options, memProvider, marker))
}
