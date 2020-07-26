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
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
)

// PackageRevisionDesiredState is the desired state of the package revision.
type PackageRevisionDesiredState string

const (
	// PackageRevisionActive is an active package revision.
	PackageRevisionActive PackageRevisionDesiredState = "Active"

	// PackageRevisionInactive is an inactive package revision.
	PackageRevisionInactive PackageRevisionDesiredState = "Inactive"
)

// +kubebuilder:object:root=true

// A PackageRevision that has been added to Crossplane.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditionedStatus.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="IMAGE",type="string",JSONPath=".spec.image"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".spec.desiredState"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane}
type PackageRevision struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PackageRevisionSpec   `json:"spec,omitempty"`
	Status PackageRevisionStatus `json:"status,omitempty"`
}

// PackageRevisionSpec specifies the desired state of a PackageRevision.
type PackageRevisionSpec struct {
	CRDs                       []metav1.TypeMeta           `json:"customresourcedefinitions,omitempty"`
	InfrastructureDefinitions  []metav1.TypeMeta           `json:"infrastructuredefinitions,omitempty"`
	InfrastructurePublications []metav1.TypeMeta           `json:"infrastructurepublications,omitempty"`
	Compositions               []metav1.TypeMeta           `json:"compositions,omitempty"`
	Controller                 ControllerSpec              `json:"controller,omitempty"`
	InstallJobRef              *corev1.ObjectReference     `json:"installJobRef,omitempty"`
	DesiredState               PackageRevisionDesiredState `json:"desiredState"`
	Image                      string                      `json:"image"`
	Revision                   int64                       `json:"revision"`
	// DependsOn is the list of packages and CRDs that this package depends on.
	DependsOn []Dependency `json:"dependsOn,omitempty"`
}

// Dependency specifies the dependency of a package.
type Dependency struct {
	// Package is the name of the package package that is being requested, e.g.,
	// myapp. Either Package or CustomResourceDefinition can be specified.
	Package string `json:"package,omitempty"`

	// CustomResourceDefinition is the full name of a CRD that is owned by the
	// package being requested. This can be a convenient way of installing a
	// package when the desired CRD is known, but the package name that contains
	// it is not known. Either Package or CustomResourceDefinition can be
	// specified.
	CustomResourceDefinition string `json:"crd,omitempty"`
}

// ControllerSpec defines the controller that implements the logic for a
// package, which can come in different flavors.
type ControllerSpec struct {
	// ServiceAccount options allow for changes to the ServiceAccount the
	// Package Manager creates for the Package's controller
	ServiceAccount *ServiceAccountOptions `json:"serviceAccount,omitempty"`

	Deployment *ControllerDeployment `json:"deployment,omitempty"`
}

// PermissionsSpec defines the permissions that a package will require to
// operate.
type PermissionsSpec struct {
	Rules []rbac.PolicyRule `json:"rules,omitempty"`
}

// ContributorSpec defines a contributor for a package (e.g., maintainer, owner,
// etc.)
type ContributorSpec struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

// ControllerDeployment defines a controller for a package that is managed by a
// Deployment.
type ControllerDeployment struct {
	Name string              `json:"name"`
	Spec apps.DeploymentSpec `json:"spec"`
}

// PackageRevisionStatus represents the observed state of a PackageRevision.
type PackageRevisionStatus struct {
	runtimev1alpha1.ConditionedStatus `json:"conditionedStatus,omitempty"`
	ControllerRef                     *corev1.ObjectReference `json:"controllerRef,omitempty"`
}

// +kubebuilder:object:root=true

// PackageRevisionList contains a list of Package.
type PackageRevisionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Package `json:"items"`
}
