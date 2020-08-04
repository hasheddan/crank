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
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
)

var _ Package = &Provider{}
var _ Package = &Configuration{}

// Package is the interface satisfied by package types.
// +k8s:deepcopy-gen=false
type Package interface {
	resource.Object
	resource.Conditioned

	GetSource() string
	SetSource(s string)

	GetCurrentRevision() string
	SetCurrentRevision(r string)
}

// GetCondition of this Provider.
func (p *Provider) GetCondition(ct runtimev1alpha1.ConditionType) runtimev1alpha1.Condition {
	return p.Status.GetCondition(ct)
}

// SetConditions of this Provider.
func (p *Provider) SetConditions(c ...runtimev1alpha1.Condition) {
	p.Status.SetConditions(c...)
}

// GetSource of this Provider.
func (p *Provider) GetSource() string {
	return p.Spec.Package
}

// SetSource of this Provider.
func (p *Provider) SetSource(s string) {
	p.Spec.Package = s
}

// GetCurrentRevision of this Provider.
func (p *Provider) GetCurrentRevision() string {
	return p.Status.CurrentRevision
}

// SetCurrentRevision of this Provider.
func (p *Provider) SetCurrentRevision(s string) {
	p.Status.CurrentRevision = s
}

// GetCondition of this Configuration.
func (p *Configuration) GetCondition(ct runtimev1alpha1.ConditionType) runtimev1alpha1.Condition {
	return p.Status.GetCondition(ct)
}

// SetConditions of this Configuration.
func (p *Configuration) SetConditions(c ...runtimev1alpha1.Condition) {
	p.Status.SetConditions(c...)
}

// GetSource of this Configuration.
func (p *Configuration) GetSource() string {
	return p.Spec.Package
}

// SetSource of this Configuration.
func (p *Configuration) SetSource(s string) {
	p.Spec.Package = s
}

// GetCurrentRevision of this Configuration.
func (p *Configuration) GetCurrentRevision() string {
	return p.Status.CurrentRevision
}

// SetCurrentRevision of this Configuration.
func (p *Configuration) SetCurrentRevision(s string) {
	p.Status.CurrentRevision = s
}

var _ PackageRevision = &ProviderRevision{}
var _ PackageRevision = &ConfigurationRevision{}

// PackageRevision is the interface satisfied by package revision types.
// +k8s:deepcopy-gen=false
type PackageRevision interface {
	resource.Object
	resource.Conditioned

	GetSource() string
	SetSource(s string)

	GetDesiredState() PackageRevisionDesiredState
	SetDesiredState(d PackageRevisionDesiredState)

	GetRevision() int64
	SetRevision(r int64)
}

// GetCondition of this ProviderRevision.
func (p *ProviderRevision) GetCondition(ct runtimev1alpha1.ConditionType) runtimev1alpha1.Condition {
	return p.Status.GetCondition(ct)
}

// SetConditions of this ProviderRevision.
func (p *ProviderRevision) SetConditions(c ...runtimev1alpha1.Condition) {
	p.Status.SetConditions(c...)
}

// GetSource of this ProviderRevision.
func (p *ProviderRevision) GetSource() string {
	return p.Spec.Image
}

// SetSource of this ProviderRevision.
func (p *ProviderRevision) SetSource(s string) {
	p.Spec.Image = s
}

// GetDesiredState of this ProviderRevision.
func (p *ProviderRevision) GetDesiredState() PackageRevisionDesiredState {
	return p.Spec.DesiredState
}

// SetDesiredState of this ProviderRevision.
func (p *ProviderRevision) SetDesiredState(s PackageRevisionDesiredState) {
	p.Spec.DesiredState = s
}

// GetRevision of this ProviderRevision.
func (p *ProviderRevision) GetRevision() int64 {
	return p.Spec.Revision
}

// SetRevision of this ProviderRevision.
func (p *ProviderRevision) SetRevision(r int64) {
	p.Spec.Revision = r
}

// GetCondition of this ConfigurationRevision.
func (p *ConfigurationRevision) GetCondition(ct runtimev1alpha1.ConditionType) runtimev1alpha1.Condition {
	return p.Status.GetCondition(ct)
}

// SetConditions of this ConfigurationRevision.
func (p *ConfigurationRevision) SetConditions(c ...runtimev1alpha1.Condition) {
	p.Status.SetConditions(c...)
}

// GetSource of this ConfigurationRevision.
func (p *ConfigurationRevision) GetSource() string {
	return p.Spec.Image
}

// SetSource of this ConfigurationRevision.
func (p *ConfigurationRevision) SetSource(s string) {
	p.Spec.Image = s
}

// GetDesiredState of this ConfigurationRevision.
func (p *ConfigurationRevision) GetDesiredState() PackageRevisionDesiredState {
	return p.Spec.DesiredState
}

// SetDesiredState of this ConfigurationRevision.
func (p *ConfigurationRevision) SetDesiredState(s PackageRevisionDesiredState) {
	p.Spec.DesiredState = s
}

// GetRevision of this ConfigurationRevision.
func (p *ConfigurationRevision) GetRevision() int64 {
	return p.Spec.Revision
}

// SetRevision of this ConfigurationRevision.
func (p *ConfigurationRevision) SetRevision(r int64) {
	p.Spec.Revision = r
}
