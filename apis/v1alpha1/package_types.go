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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
)

// +kubebuilder:object:root=true

// Package is the CRD type for a request to add a package to Crossplane.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditionedStatus.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SOURCE",type="string",JSONPath=".spec.source"
// +kubebuilder:printcolumn:name="PACKAGE",type="string",JSONPath=".spec.package"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane}
type Package struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PackageSpec   `json:"spec,omitempty"`
	Status PackageStatus `json:"status,omitempty"`
}

// PackageSpec specifies details about a request to install a package to
// Crossplane.
type PackageSpec struct {
	PackageControllerOptions `json:",inline"`

	// Package is the name of the package that is being requested, e.g.,
	// myapp. Either Package or CustomResourceDefinition can be specified.
	Package string `json:"package,omitempty"`
}

// PackageControllerOptions allow for changes in the Package extraction and
// deployment controllers. These can affect how images are fetched and how
// Package derived resources are created.
type PackageControllerOptions struct {
	// ImagePullSecrets are named secrets in the same workspace that can be used
	// to fetch Packages from private repositories and to run controllers from
	// private repositories
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// ImagePullPolicy defines the pull policy for all images used during
	// Package extraction and when running the Package controller.
	// https://kubernetes.io/docs/concepts/configuration/overview/#container-images
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// ServiceAccount options allow for changes to the ServiceAccount the
	// Package Manager creates for the PackageRevision's controller
	ServiceAccount *ServiceAccountOptions `json:"serviceAccount,omitempty"`
}

// ServiceAccountOptions augment the ServiceAccount created by the Package
// controller
type ServiceAccountOptions struct {
	Annotations map[string]string `json:"annotations,omitempty"`
}

// PackageStatus represents the observed state of a Package.
type PackageStatus struct {
	runtimev1alpha1.ConditionedStatus `json:"conditionedStatus,omitempty"`

	CurrentRevision string `json:"currentRevision,omitempty"`
}

// +kubebuilder:object:root=true

// PackageList contains a list of Package.
type PackageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Package `json:"items"`
}
