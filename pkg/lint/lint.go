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

package lint

import (
	"fmt"
	"strings"

	"github.com/hasheddan/crank/pkg/parser"
)

// Linter lints Crossplane packages.
type Linter struct {
	pkg     *parser.Package
	verbose bool
}

// NewLinter creates a new linter for the file system.
func NewLinter(p *parser.Package) *Linter {
	return &Linter{
		pkg:     p,
		verbose: false,
	}
}

// Lint executes the linters.
// TODO(hasheddan): should be able to supply pluggable linters.
func (l *Linter) Lint() []string {
	errors := []string{}
	for _, c := range l.pkg.Compositions {
		if _, ok := l.pkg.InfrastructureDefinitions[strings.ToLower(c.Spec.From.Kind)]; !ok {
			errors = append(errors, fmt.Sprintf("Composition %s satisfies a Definition that does not exist: %s", c.Name, c.Spec.From.Kind))
		}
	}
	return errors
}
