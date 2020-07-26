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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true

// Module is the CRD type for a request to add a package to Crossplane.
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane}
type Module struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ModuleSpec `json:"spec,omitempty"`
}

// ModuleSpec specifies details about a request to install a package to
// Crossplane.
type ModuleSpec struct {
	Packages                  map[string]PackageDependencies `json:"packages"`
	Compositions              map[string]string              `json:"compositions,omitempty"`
	CustomResourceDefinitions map[string]string              `json:"crds,omitempty"`
}

// PackageDependencies identifies the dependencies of a package.
type PackageDependencies struct {
	Name         string   `json:"name,omitempty"`
	Image        string   `json:"image,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
}

// +kubebuilder:object:root=true

// ModuleList contains a list of Module.
type ModuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Module `json:"items"`
}
