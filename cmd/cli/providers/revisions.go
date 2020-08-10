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

package providers

import (
	"context"
	"fmt"

	"github.com/hasheddan/crank/apis"
	"github.com/hasheddan/crank/apis/v1alpha1"
	"github.com/hasheddan/crank/pkg/prompt"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// revisions will list revisions for a Provider.
var revisions = &cobra.Command{
	Use:   "revisions",
	Short: "List all revisions for an installed Provider",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := ctrl.GetConfig()
		if err != nil {
			panic(err)
		}
		s := runtime.NewScheme()
		if err := apis.AddToScheme(s); err != nil {
			panic(err)
		}
		c, err := client.New(conf, client.Options{Scheme: s})
		if err != nil {
			panic(err)
		}
		prs := &v1alpha1.ProviderRevisionList{}
		if err := c.List(context.TODO(), prs, client.MatchingLabels(map[string]string{"crank.crossplane.io/package": args[0]})); err != nil {
			panic(err)
		}
		for _, p := range prs.Items {
			fmt.Printf(prompt.FmtInfo(p.Name))
		}
	},
}
