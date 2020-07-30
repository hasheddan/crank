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

package packages

import (
	"context"
	"fmt"

	"github.com/hasheddan/crank/apis"
	"github.com/hasheddan/crank/apis/v1alpha1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ctrl "sigs.k8s.io/controller-runtime"
)

// list will list all installed Crossplane package.
var list = &cobra.Command{
	Use:   "list",
	Short: "List all installed Crossplane package",
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
		m := &v1alpha1.PackageLock{}
		if err := c.Get(context.TODO(), types.NamespacedName{Name: "packages"}, m); err != nil {
			panic(err)
		}
		for n, p := range m.Spec.Packages {
			fmt.Printf(InfoColor, n)
			fmt.Printf(NoticeColor, "  📦 "+p.Image)
			for _, d := range p.Dependencies {
				fmt.Printf(NoticeColor, " 👉 "+d)
			}
		}
	},
}
