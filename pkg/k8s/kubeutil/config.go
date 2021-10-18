/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubeutil

import (
	"errors"
	"github.com/crain-cn/event-mesh/pkg/logging"
	"github.com/crain-cn/event-mesh/pkg/logging/logfields"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	log = logging.DefaultLogger.WithField(logfields.LogSubsys, "kubeutil")
)

func LoadConfig(kubeconfig string) (*rest.Config, error) {
	cfg, err := clientcmd.Load([]byte(kubeconfig))
	if err != nil {
		log.WithError(err).Fatalf("unmarshal: %v", err)
	}
	for context := range cfg.Contexts {
		log.Infof("* %s", context)
		contextCfg, err := clientcmd.NewNonInteractiveClientConfig(*cfg, context, &clientcmd.ConfigOverrides{}, nil).ClientConfig()
		if err != nil {
			logrus.WithError(err).Fatalf("create %s client: %v", context, err)
		}
		// An arbitrary high number we expect to not exceed. There are various components that need more than the default 5 QPS/10 Burst, e.G.
		// hook for creating ProwJobs and Plank for creating Pods.
		contextCfg.QPS = 100
		contextCfg.Burst = 1000
		return contextCfg, nil
	}
	return nil, errors.New("invalid kubeconfig")
}
