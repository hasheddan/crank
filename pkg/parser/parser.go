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

package parser

import (
	"bufio"
	"errors"
	"os"
	"strings"

	apiv1alpha1 "github.com/crossplane/crossplane/apis/apiextensions/v1alpha1"
	"github.com/ghodss/yaml"
	"github.com/hasheddan/crank/apis/v1alpha1"
	"github.com/spf13/afero"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

// Package is a Crossplane package.
type Package struct {
	Name                       string
	CustomResourceDefinitions  map[string]apiextensions.CustomResourceDefinition
	InfrastructureDefinitions  map[string]apiv1alpha1.InfrastructureDefinition
	InfrastructurePublications map[string]apiv1alpha1.InfrastructurePublication
	Compositions               map[string]apiv1alpha1.Composition
	Dependencies               []v1alpha1.Dependency
}

// Parser parses a package.
type Parser struct {
	fs afero.Fs
}

// NewParser constructs a new parser.
func NewParser(fs afero.Fs) *Parser {
	return &Parser{
		fs: fs,
	}
}

// ParsePackage parses a package at the given path and returns it.
func (p *Parser) ParsePackage(root string) (*Package, error) {
	pkg := &Package{
		CustomResourceDefinitions:  map[string]apiextensions.CustomResourceDefinition{},
		InfrastructureDefinitions:  map[string]apiv1alpha1.InfrastructureDefinition{},
		InfrastructurePublications: map[string]apiv1alpha1.InfrastructurePublication{},
		Compositions:               map[string]apiv1alpha1.Composition{},
	}
	if err := afero.Walk(p.fs, root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		b, err := afero.ReadFile(p.fs, path)
		if err != nil {
			return err
		}
		crd := &apiextensions.CustomResourceDefinition{}
		if err := parseCRD(b, crd); err == nil {
			pkg.CustomResourceDefinitions[path] = *crd
			return nil
		}
		id := &apiv1alpha1.InfrastructureDefinition{}
		if err := parseID(b, id); err == nil {
			pkg.InfrastructureDefinitions[path] = *id
			return nil
		}
		ip := &apiv1alpha1.InfrastructurePublication{}
		if err := parseIP(b, ip); err == nil {
			pkg.InfrastructurePublications[path] = *ip
			return nil
		}
		c := &apiv1alpha1.Composition{}
		if err := parseC(b, c); err == nil {
			pkg.Compositions[path] = *c
			return nil
		}
		return nil
	}); err != nil {
		return pkg, err
	}
	return pkg, nil
}

// ParseLines finds the start and end line for the match.
func (p *Parser) ParseLines(path, startMatch, endMatch string) (int, int, error) {
	var startLine, endLine int
	f, err := p.fs.Open(path)
	if err != nil {
		return startLine, endLine, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	line := 0
	foundStart := false
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), startMatch) {
			foundStart = true
			startLine = line
		}
		if strings.Contains(scanner.Text(), endMatch) && foundStart {
			endLine = line
			break
		}
		line++
	}
	err = scanner.Err()
	return startLine, endLine, err
}

func parseCRD(b []byte, crd *apiextensions.CustomResourceDefinition) error {
	err := yaml.Unmarshal(b, crd)
	if err == nil && crd.Kind != "CustomResourceDefinition" {
		return errors.New("not a CustomResourceDefintion")
	}
	return err
}

func parseID(b []byte, id *apiv1alpha1.InfrastructureDefinition) error {
	err := yaml.Unmarshal(b, id)
	if err == nil && id.Kind != apiv1alpha1.InfrastructureDefinitionKind {
		return errors.New("not an InfrastructureDefinition")
	}
	return err
}

func parseIP(b []byte, ip *apiv1alpha1.InfrastructurePublication) error {
	err := yaml.Unmarshal(b, ip)
	if err == nil && ip.Kind != apiv1alpha1.InfrastructurePublicationKind {
		return errors.New("not an InfrastructurePublication")
	}
	return err
}

func parseC(b []byte, c *apiv1alpha1.Composition) error {
	err := yaml.Unmarshal(b, c)
	if err == nil && c.Kind != apiv1alpha1.CompositionKind {
		return errors.New("not a Composition")
	}
	return err
}
