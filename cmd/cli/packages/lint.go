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
	"fmt"
	"path/filepath"

	"github.com/hasheddan/crank/pkg/lint"
	"github.com/hasheddan/crank/pkg/parser"
	"github.com/hasheddan/crank/pkg/prompt"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// linter will lint a Crossplane package.
var linter = &cobra.Command{
	Use:   "lint",
	Short: "Lints a Crossplane package",
	Long:  `Linting a Crossplane package ensures that the format is suitable for building.`,
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) == 1 {
			path = args[0]
		}
		s, _ := filepath.Abs(path)
		fs := afero.NewOsFs()
		p := parser.NewParser(fs)
		pkg, err := p.ParsePackage(s)
		if err != nil {
			panic(err)
		}
		l := lint.NewLinter(pkg)
		errs := l.Lint()
		errLen := len(errs)
		if errLen != 0 {
			fmt.Printf(prompt.FmtWarning(fmt.Sprintf("Found %d errors in package.\n", errLen)))
		}
		for i, e := range errs {
			fmt.Printf(prompt.FmtError(fmt.Sprintf("[%d/%d] %s\n", i+1, errLen, e)))
		}
	},
}
