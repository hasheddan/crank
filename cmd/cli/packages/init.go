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
	"bufio"
	"fmt"
	"os"
	"text/template"

	"github.com/hasheddan/crank/pkg/parser"
	"github.com/hasheddan/crank/pkg/prompt"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var (
	pkg *parser.Package
)

// CrossplaneYAML is the crossplane.yaml template.
var CrossplaneYAML = template.Must(template.New("crossplane").Parse(
	`apiVersion: pkg.crossplane.io/v1alpha1
kind: {{ .PackageType }}
metadata:
  name: {{ .Name }}
  annotations:
spec:
  dependsOn:{{ range $i, $dep := .Dependencies }}
  - package: {{ $dep.Package }}
    version: {{ $dep.Version }}{{ end }}
  ignore:
  - path: examples/
`))

// PackageType is a type of Crossplane package.
type PackageType string

// PackageType is either Configuration or Provider.
const (
	Configuration PackageType = "Configuration"
	Provider      PackageType = "Provider"
)

// Steps in prompter execution.
const (
	First prompt.Step = iota
	Second
	Third
	Fourth
)

type initer struct {
	PackageType  PackageType
	Name         string
	Dependencies []Dependency
}

// A Dependency is a dependency of a package.
type Dependency struct {
	Package string
	Version string
}

var i = &initer{}

// Init Prompter
var p = prompt.NewPrompter("What type of Package would you like to init?", bufio.NewScanner(os.Stdin), map[prompt.Step]prompt.Fn{
	First:  firstStep,
	Second: secondStep,
	Third:  prompt.BinaryFn(thirdStep),
	Fourth: prompt.BareFn(fourthStep),
}, First)

var depP = prompt.NewPrompter("What is the package dependency name?", bufio.NewScanner(os.Stdin), map[prompt.Step]prompt.Fn{
	First:  depStepOne,
	Second: depStepTwo,
}, First)

func firstStep(p prompt.Prompter, s string) (string, bool) {
	if s != "Configuration" && s != "Provider" {
		return "Package must be of type Configuration or Provider.", false
	}
	i.PackageType = PackageType(s)
	p.SetStep(Second)
	return fmt.Sprintf("What would you like to name your %s?", s), false
}

func secondStep(p prompt.Prompter, s string) (string, bool) {
	i.Name = s
	p.SetStep(Third)
	return fmt.Sprintf("Does %s have any dependencies? (Y/N)", s), false
}

func thirdStep(b bool, p prompt.Prompter, s string) (string, bool) {
	if !b {
		p.SetStep(Fourth)
		return fmt.Sprintf("Package of type %s will be initialized with name %s. Ok? (Y/N)", i.PackageType, i.Name), false
	}
	fmt.Println(prompt.FmtInfo("Enter 'done' when finished adding dependencies."))
	if err := depP.Prompt(); err != nil {
		panic(err)
	}
	p.SetStep(Fourth)
	return fmt.Sprintf("Package of type %s will be initialized with name %s. Ok? (Y/N)", i.PackageType, i.Name), false
}

func depStepOne(p prompt.Prompter, s string) (string, bool) {
	if s == "done" {
		return prompt.FmtInfo(fmt.Sprintf("Added %d dependencies.", len(i.Dependencies))), true
	}
	i.Dependencies = append(i.Dependencies, Dependency{Package: s})
	p.SetStep(Second)
	return fmt.Sprintf("What version of package %s is required?", s), false
}

func depStepTwo(p prompt.Prompter, s string) (string, bool) {
	i.Dependencies[len(i.Dependencies)-1].Version = s
	p.SetStep(First)
	return "What is the package dependency name?", false
}

func fourthStep() (string, bool) {
	fmt.Printf(prompt.FmtInfo("📦 Building package...\n"))
	if err := CrossplaneYAML.Execute(os.Stdout, i); err != nil {
		return "Failed to initialize package.", true
	}
	return prompt.FmtInfo("✔️  Package initialized."), true
}

// initialize will init a Crossplane package.
var initialize = &cobra.Command{
	Use:   "init",
	Short: "Initializes a Crossplane package",
	Long:  `Initializes a Crossplane package.`,
	Run: func(cmd *cobra.Command, args []string) {
		fs := afero.NewOsFs()
		parse := parser.NewParser(fs)
		root, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		pkg, err = parse.ParsePackage(root)
		if err != nil {
			panic(err)
		}
		if len(pkg.InfrastructureDefinitions) > 0 {
			fmt.Println(prompt.FmtNotice(fmt.Sprintf("Found %d InfrastructureDefinitions:", len(pkg.InfrastructureDefinitions))))
		}
		for _, id := range pkg.InfrastructureDefinitions {
			fmt.Println(prompt.FmtInfo(fmt.Sprintf("-- %s", id.Name)))
		}
		if len(pkg.InfrastructurePublications) > 0 {
			fmt.Println(prompt.FmtNotice(fmt.Sprintf("Found %d InfrastructurePublications:", len(pkg.InfrastructurePublications))))
		}
		for _, ip := range pkg.InfrastructurePublications {
			fmt.Println(prompt.FmtInfo(fmt.Sprintf("-- %s", ip.Name)))
		}
		if len(pkg.Compositions) > 0 {
			fmt.Println(prompt.FmtNotice(fmt.Sprintf("Found %d Compositions:", len(pkg.Compositions))))
		}
		for _, c := range pkg.Compositions {
			fmt.Println(prompt.FmtInfo(fmt.Sprintf("-- %s", c.Name)))
		}
		if err := p.Prompt(); err != nil {
			panic(err)
		}
	},
}
