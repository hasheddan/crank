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
	"github.com/ghodss/yaml"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/hasheddan/crank/apis/v1alpha1"
	"github.com/hasheddan/veneer"
	"github.com/spf13/afero"
)

// Unpack unpacks and image and gets its dependencies.
func Unpack(image string) ([]string, error) {
	ref, err := name.ParseReference(image)
	if err != nil {
		return nil, err
	}

	img, err := remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return nil, err
	}

	layers, err := img.Layers()
	if err != nil {
		return nil, err
	}

	fs := afero.NewMemMapFs()
	err = veneer.LayerFs(layers[len(layers)-1], fs)
	if err != nil {
		return nil, err
	}
	bytes, err := afero.ReadFile(fs, ".registry/app.yaml")
	if err != nil {
		return nil, err
	}
	deps := &AppMetadataSpec{}
	err = yaml.Unmarshal(bytes, deps)
	if err != nil {
		return nil, err
	}
	stringDeps := make([]string, len(deps.DependsOn))
	for i, d := range deps.DependsOn {
		stringDeps[i] = d.Package
	}
	return stringDeps, nil
}

// AppMetadataSpec defines metadata about the package application
type AppMetadataSpec struct {
	DependsOn []v1alpha1.Dependency `json:"dependsOn,omitempty"`
}
