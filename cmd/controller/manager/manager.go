/*
Copyright 2020 The Crossplane Authors.

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

package manager

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/pkg/errors"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/hasheddan/crank/apis"
	"github.com/hasheddan/crank/pkg/controller"
)

var (
	sync  = time.Hour
	debug = true
)

// Root is the root command for the manager.
var Root = &cobra.Command{
	Use:   "manager",
	Short: "Run the crank package manager",
	Long:  `The crank package manager is meant to run in a Kubernetes cluster where Crossplane is installed.`,
	Run: func(cmd *cobra.Command, args []string) {
		zl := zap.New(zap.UseDevMode(debug))
		if debug {
			// The controller-runtime runs with a no-op logger by default. It is
			// *very* verbose even at info level, so we only provide it a real
			// logger when we're running in debug mode.
			ctrl.SetLogger(zl)
		}
		if err := run(logging.NewLogrLogger(zl.WithName("manager"))); err != nil {
			panic(err)
		}
	},
}

func run(log logging.Logger) error {
	log.Debug("Starting", "sync-period", sync.String()) // TODO
	cfg, err := getRestConfig("")                       // TODO
	if err != nil {
		return errors.Wrap(err, "Cannot get config")
	}

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{SyncPeriod: &sync}) // TODO
	if err != nil {
		return errors.Wrap(err, "Cannot create manager")
	}

	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		return errors.Wrap(err, "Cannot add core Crossplane APIs to scheme")
	}

	if err := apiextensionsv1beta1.AddToScheme(mgr.GetScheme()); err != nil {
		return errors.Wrap(err, "Cannot add API extensions to scheme")
	}

	if err := controller.Setup(mgr, log); err != nil {
		return errors.Wrap(err, "Cannot setup package manager controllers")
	}

	return errors.Wrap(mgr.Start(ctrl.SetupSignalHandler()), "Cannot start controller manager")
}

func getRestConfig(kubeconfigPath string) (*rest.Config, error) {
	if kubeconfigPath == "" {
		return ctrl.GetConfig()
	}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{}).ClientConfig()
}
