// +build !ignore_autogenerated

/*
Copyright 2020 The Knative Authors

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

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AdditionalProperties) DeepCopyInto(out *AdditionalProperties) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AdditionalProperties.
func (in *AdditionalProperties) DeepCopy() *AdditionalProperties {
	if in == nil {
		return nil
	}
	out := new(AdditionalProperties)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DomainExtractionFilterStrategySpec) DeepCopyInto(out *DomainExtractionFilterStrategySpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DomainExtractionFilterStrategySpec.
func (in *DomainExtractionFilterStrategySpec) DeepCopy() *DomainExtractionFilterStrategySpec {
	if in == nil {
		return nil
	}
	out := new(DomainExtractionFilterStrategySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DomainExtractionGroupingStrategySpec) DeepCopyInto(out *DomainExtractionGroupingStrategySpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DomainExtractionGroupingStrategySpec.
func (in *DomainExtractionGroupingStrategySpec) DeepCopy() *DomainExtractionGroupingStrategySpec {
	if in == nil {
		return nil
	}
	out := new(DomainExtractionGroupingStrategySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DomainExtractionParametersSpec) DeepCopyInto(out *DomainExtractionParametersSpec) {
	*out = *in
	out.DomainExtractionStrategies = in.DomainExtractionStrategies
	out.DomainSchemaRegistryProps = in.DomainSchemaRegistryProps
	out.AdditionalProperties = in.AdditionalProperties
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DomainExtractionParametersSpec.
func (in *DomainExtractionParametersSpec) DeepCopy() *DomainExtractionParametersSpec {
	if in == nil {
		return nil
	}
	out := new(DomainExtractionParametersSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DomainExtractionStrategiesSpec) DeepCopyInto(out *DomainExtractionStrategiesSpec) {
	*out = *in
	out.FilterStrategy = in.FilterStrategy
	out.GroupingStrategy = in.GroupingStrategy
	out.UseJobTimestampStrategy = in.UseJobTimestampStrategy
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DomainExtractionStrategiesSpec.
func (in *DomainExtractionStrategiesSpec) DeepCopy() *DomainExtractionStrategiesSpec {
	if in == nil {
		return nil
	}
	out := new(DomainExtractionStrategiesSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DomainExtractionUseJobTsSpec) DeepCopyInto(out *DomainExtractionUseJobTsSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DomainExtractionUseJobTsSpec.
func (in *DomainExtractionUseJobTsSpec) DeepCopy() *DomainExtractionUseJobTsSpec {
	if in == nil {
		return nil
	}
	out := new(DomainExtractionUseJobTsSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DomainSchemaRegistrySpec) DeepCopyInto(out *DomainSchemaRegistrySpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DomainSchemaRegistrySpec.
func (in *DomainSchemaRegistrySpec) DeepCopy() *DomainSchemaRegistrySpec {
	if in == nil {
		return nil
	}
	out := new(DomainSchemaRegistrySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FrameworkParametersSpec) DeepCopyInto(out *FrameworkParametersSpec) {
	*out = *in
	in.BasicAuthUser.DeepCopyInto(&out.BasicAuthUser)
	in.BasicAuthPassword.DeepCopyInto(&out.BasicAuthPassword)
	in.OAuthURL.DeepCopyInto(&out.OAuthURL)
	in.OAuthScpUser.DeepCopyInto(&out.OAuthScpUser)
	in.OAuthScpPassword.DeepCopyInto(&out.OAuthScpPassword)
	in.OAuthScpClientID.DeepCopyInto(&out.OAuthScpClientID)
	in.OAuthScpClientSecret.DeepCopyInto(&out.OAuthScpClientSecret)
	in.EventLogEndpoint.DeepCopyInto(&out.EventLogEndpoint)
	in.EventLogUser.DeepCopyInto(&out.EventLogUser)
	in.EventLogPassword.DeepCopyInto(&out.EventLogPassword)
	in.SchemaRegistryEndpoint.DeepCopyInto(&out.SchemaRegistryEndpoint)
	in.SchemaRegistryCreds.DeepCopyInto(&out.SchemaRegistryCreds)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FrameworkParametersSpec.
func (in *FrameworkParametersSpec) DeepCopy() *FrameworkParametersSpec {
	if in == nil {
		return nil
	}
	out := new(FrameworkParametersSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IntegrationScenario) DeepCopyInto(out *IntegrationScenario) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IntegrationScenario.
func (in *IntegrationScenario) DeepCopy() *IntegrationScenario {
	if in == nil {
		return nil
	}
	out := new(IntegrationScenario)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *IntegrationScenario) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IntegrationScenarioLimitsSpec) DeepCopyInto(out *IntegrationScenarioLimitsSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IntegrationScenarioLimitsSpec.
func (in *IntegrationScenarioLimitsSpec) DeepCopy() *IntegrationScenarioLimitsSpec {
	if in == nil {
		return nil
	}
	out := new(IntegrationScenarioLimitsSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IntegrationScenarioList) DeepCopyInto(out *IntegrationScenarioList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]IntegrationScenario, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IntegrationScenarioList.
func (in *IntegrationScenarioList) DeepCopy() *IntegrationScenarioList {
	if in == nil {
		return nil
	}
	out := new(IntegrationScenarioList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *IntegrationScenarioList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IntegrationScenarioRequestsSpec) DeepCopyInto(out *IntegrationScenarioRequestsSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IntegrationScenarioRequestsSpec.
func (in *IntegrationScenarioRequestsSpec) DeepCopy() *IntegrationScenarioRequestsSpec {
	if in == nil {
		return nil
	}
	out := new(IntegrationScenarioRequestsSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IntegrationScenarioResourceSpec) DeepCopyInto(out *IntegrationScenarioResourceSpec) {
	*out = *in
	out.Requests = in.Requests
	out.Limits = in.Limits
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IntegrationScenarioResourceSpec.
func (in *IntegrationScenarioResourceSpec) DeepCopy() *IntegrationScenarioResourceSpec {
	if in == nil {
		return nil
	}
	out := new(IntegrationScenarioResourceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IntegrationScenarioSpec) DeepCopyInto(out *IntegrationScenarioSpec) {
	*out = *in
	out.DomainExtractionParameters = in.DomainExtractionParameters
	in.FrameworkParameters.DeepCopyInto(&out.FrameworkParameters)
	out.Resources = in.Resources
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IntegrationScenarioSpec.
func (in *IntegrationScenarioSpec) DeepCopy() *IntegrationScenarioSpec {
	if in == nil {
		return nil
	}
	out := new(IntegrationScenarioSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IntegrationScenarioStatus) DeepCopyInto(out *IntegrationScenarioStatus) {
	*out = *in
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IntegrationScenarioStatus.
func (in *IntegrationScenarioStatus) DeepCopy() *IntegrationScenarioStatus {
	if in == nil {
		return nil
	}
	out := new(IntegrationScenarioStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretValueFromSource) DeepCopyInto(out *SecretValueFromSource) {
	*out = *in
	if in.SecretKeyRef != nil {
		in, out := &in.SecretKeyRef, &out.SecretKeyRef
		*out = new(v1.SecretKeySelector)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretValueFromSource.
func (in *SecretValueFromSource) DeepCopy() *SecretValueFromSource {
	if in == nil {
		return nil
	}
	out := new(SecretValueFromSource)
	in.DeepCopyInto(out)
	return out
}
