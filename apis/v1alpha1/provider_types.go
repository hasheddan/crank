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

// Provider is the CRD type for a request to add a provider to Crossplane.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditionedStatus.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="PACKAGE",type="string",JSONPath=".spec.package"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane}
type Provider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProviderSpec   `json:"spec,omitempty"`
	Status ProviderStatus `json:"status,omitempty"`
}

// ProviderSpec specifies details about a request to install a provider to
// Crossplane.
type ProviderSpec struct {
	ProviderControllerOptions `json:",inline"`

	// Package is the name of the package that is being requested, e.g., myapp.
	Package string `json:"package"`
}

// ProviderControllerOptions allow for changes in the Provider extraction and
// deployment controllers. These can affect how images are fetched and how
// Provider derived resources are created.
type ProviderControllerOptions struct {
	// ImagePullSecrets are named secrets in the same workspace that can be used
	// to fetch Providers from private repositories and to run controllers from
	// private repositories
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// ImagePullPolicy defines the pull policy for all images used during
	// Provider extraction and when running the Provider controller.
	// https://kubernetes.io/docs/concepts/configuration/overview/#container-images
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// ServiceAccount options allow for changes to the ServiceAccount the
	// Package Manager creates for the Provider's controller
	ServiceAccount *ServiceAccountOptions `json:"serviceAccount,omitempty"`
}

// ServiceAccountOptions augment the ServiceAccount created by the
// PackageRevision controller for the Provider
type ServiceAccountOptions struct {
	Annotations map[string]string `json:"annotations,omitempty"`
}

// ProviderStatus represents the observed state of a Provider.
type ProviderStatus struct {
	runtimev1alpha1.ConditionedStatus `json:"conditionedStatus,omitempty"`

	CurrentRevision string `json:"currentRevision,omitempty"`
}

// +kubebuilder:object:root=true

// ProviderList contains a list of Provider.
type ProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Provider `json:"items"`
}

// +kubebuilder:object:root=true

// A ProviderRevision that has been added to Crossplane.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditionedStatus.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="IMAGE",type="string",JSONPath=".spec.image"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".spec.desiredState"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane}
type ProviderRevision struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PackageRevisionSpec   `json:"spec,omitempty"`
	Status PackageRevisionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProviderRevisionList contains a list of ProviderRevision.
type ProviderRevisionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProviderRevision `json:"items"`
}
