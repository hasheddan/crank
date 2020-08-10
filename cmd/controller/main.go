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

package main

import (
	"github.com/hasheddan/crank/cmd/controller/manager"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "crank",
		Short: "The next generation package manager for Crossplane",
		Long:  `The next generation package manager for Crossplane`,
	}
)

func init() {
	rootCmd.AddCommand(manager.Root)
}

func main() {
	rootCmd.Execute()
}
