package module

import (
	"flag"
	"fmt"
	"github.com/crain-cn/event-mesh/pkg/logging"
	"github.com/sirupsen/logrus"
	"os"
)

var (
	log = logging.DefaultLogger.WithField("component", "main")
)

type options struct {
	master     string
	kubeConfig string
	configFile string
	dataDir    string
}

func ParseOptions() options {
	var o options
	if err := o.parse(flag.CommandLine, os.Args[1:]); err != nil {
		logrus.Fatalf("Invalid flags: %v", err)
	}
	return o
}

func (o *options) parse(fs *flag.FlagSet, args []string) error {
	flag.StringVar(&o.master, "master", "", "master url")
	flag.StringVar(&o.kubeConfig, "kubeconfig", "", "Path to kubeconfig. Only required if out of cluster")
	flag.StringVar(&o.configFile, "config", "config/route.yml", "")
	flag.StringVar(&o.dataDir, "data", "data/", "")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("Parse flags: %v", err)
	}
	return nil
}
