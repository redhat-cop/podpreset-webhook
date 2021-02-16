/*
Copyright 2021.

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
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PodPresetSpec defines the desired state of PodPreset
type PodPresetSpec struct {
	// +kubebuilder:validation:Required
	Selector metav1.LabelSelector `json:"selector,omitempty" protobuf:"bytes,1,opt,name=selector"`

	// +patchMergeKey=name
	// +patchStrategy=merge
	// +kubebuilder:validation:Optional
	Env []corev1.EnvVar `json:"env,omitempty" protobuf:"bytes,2,rep,name=env"`

	// +patchMergeKey=name
	// +patchStrategy=merge
	// +kubebuilder:validation:Optional
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty" protobuf:"bytes,3,rep,name=envFrom"`

	// +patchMergeKey=name
	// +patchStrategy=merge
	// +kubebuilder:validation:Optional
	Volumes []corev1.Volume `json:"volumes,omitempty" protobuf:"bytes,4,rep,name=volumes"`

	// +patchMergeKey=name
	// +patchStrategy=merge
	// +kubebuilder:validation:Optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty" protobuf:"bytes,5,rep,name=volumeMounts"`
}

// PodPresetStatus defines the observed state of PodPreset
type PodPresetStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="Conditions",xDescriptors={"urn:alm:descriptor:io.kubernetes.conditions"}
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PodPreset is the Schema for the podpresets API
type PodPreset struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PodPresetSpec   `json:"spec,omitempty"`
	Status PodPresetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PodPresetList contains a list of PodPreset
type PodPresetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PodPreset `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PodPreset{}, &PodPresetList{})
}
