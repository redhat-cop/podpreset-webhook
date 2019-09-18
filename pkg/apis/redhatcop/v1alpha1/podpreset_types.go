package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PodPresetSpec defines the desired state of PodPreset
// +k8s:openapi-gen=true
type PodPresetSpec struct {
	// Selector is a label query over a set of resources, in this case pods.
	// Required.
	Selector metav1.LabelSelector `json:"selector,omitempty" protobuf:"bytes,1,opt,name=selector"`

	// Env defines the collection of EnvVar to inject into containers.
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty" protobuf:"bytes,2,rep,name=env"`
	// EnvFrom defines the collection of EnvFromSource to inject into containers.
	// +optional
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty" protobuf:"bytes,3,rep,name=envFrom"`
	// Volumes defines the collection of Volume to inject into the pod.
	// +optional
	Volumes []corev1.Volume `json:"volumes,omitempty" protobuf:"bytes,4,rep,name=volumes"`
	// VolumeMounts defines the collection of VolumeMount to inject into containers.
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty" protobuf:"bytes,5,rep,name=volumeMounts"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodPreset is the Schema for the podpresets API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type PodPreset struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Spec PodPresetSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodPresetList contains a list of PodPreset
type PodPresetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PodPreset `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PodPreset{}, &PodPresetList{})
}
