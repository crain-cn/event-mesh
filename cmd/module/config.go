package module

import (
	"github.com/crain-cn/event-mesh/cmd/config"
	"os"
)

func ParseConfigYaml() *config.ConfigResolver {
	log.Info("ParseConfig..")
	configResolver, err := config.NewResolver("config/config.yml")
	if err != nil {
		os.Exit(1)
	}
	return configResolver
}
