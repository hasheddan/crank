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

package unpack

import (
	"os"
	"strings"

	apiv1alpha1 "github.com/crossplane/crossplane/apis/apiextensions/v1alpha1"
	"github.com/ghodss/yaml"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/hasheddan/crank/apis/v1alpha1"
	"github.com/hasheddan/veneer"
	"github.com/spf13/afero"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

// Unpack unpacks an image and gets its dependencies.
func Unpack(image string) (string, []string, error) {
	ref, err := name.ParseReference(image)
	if err != nil {
		return "", nil, err
	}

	img, err := remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return "", nil, err
	}
	hash, err := img.Digest()
	if err != nil {
		return "", nil, err
	}
	digest := strings.TrimLeft(hash.String(), "sha256:")

	layers, err := img.Layers()
	if err != nil {
		return digest, nil, err
	}

	fs := afero.NewMemMapFs()
	err = veneer.LayerFs(layers[len(layers)-1], fs)
	if err != nil {
		return digest, nil, err
	}
	bytes, err := afero.ReadFile(fs, ".registry/app.yaml")
	if err != nil {
		return digest, nil, err
	}
	deps := &AppMetadataSpec{}
	err = yaml.Unmarshal(bytes, deps)
	if err != nil {
		return digest, nil, err
	}
	stringDeps := []string{}
	for _, d := range deps.DependsOn {
		if d.Package != "" {
			stringDeps = append(stringDeps, d.Package)
		}
	}
	return digest, stringDeps, nil
}

// Resources unpacks resources from a package.
func Resources(image string) ([]apiextensions.CustomResourceDefinition, []apiv1alpha1.InfrastructureDefinition, []apiv1alpha1.Composition, error) {
	ref, err := name.ParseReference(image)
	if err != nil {
		return nil, nil, nil, err
	}

	img, err := remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return nil, nil, nil, err
	}

	layers, err := img.Layers()
	if err != nil {
		return nil, nil, nil, err
	}

	fs := afero.NewMemMapFs()
	err = veneer.LayerFs(layers[len(layers)-1], fs)
	if err != nil {
		return nil, nil, nil, err
	}
	crds := []apiextensions.CustomResourceDefinition{}
	infraDefs := []apiv1alpha1.InfrastructureDefinition{}
	comps := []apiv1alpha1.Composition{}
	if err = afero.Walk(fs, ".registry", func(path string, _ os.FileInfo, err error) error {
		b, err := afero.ReadFile(fs, path)
		if err != nil {
			return err
		}
		crd := &apiextensions.CustomResourceDefinition{}
		if err := yaml.Unmarshal(b, crd); err == nil && crd.Kind == "CustomResourceDefinition" {
			crds = append(crds, *crd)
			return nil
		}
		id := &apiv1alpha1.InfrastructureDefinition{}
		if err := yaml.Unmarshal(b, id); err == nil {
			infraDefs = append(infraDefs, *id)
			return nil
		}
		comp := &apiv1alpha1.Composition{}
		if err := yaml.Unmarshal(b, comp); err == nil {
			comps = append(comps, *comp)
			return nil
		}
		return nil
	}); err != nil {
		return nil, nil, nil, err
	}
	return crds, infraDefs, comps, nil
}

// AppMetadataSpec defines metadata about the package application
type AppMetadataSpec struct {
	DependsOn []v1alpha1.Dependency `json:"dependsOn,omitempty"`
}
