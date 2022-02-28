/*
Copyright 2022.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CustomPodSetSpec defines the desired state of CustomPodSet
type CustomPodSetSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of CustomPodSet. Edit custompodset_types.go to remove/update
	Replicas int32  `json:"replicas"`
	Image    string `json:"image"`
}

// CustomPodSetStatus defines the observed state of CustomPodSet
type CustomPodSetStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	PodNames          []string `json:"podNames"`
	AvailableReplicas int32    `json:"availableReplicas"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CustomPodSet is the Schema for the custompodsets API
type CustomPodSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomPodSetSpec   `json:"spec,omitempty"`
	Status CustomPodSetStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CustomPodSetList contains a list of CustomPodSet
type CustomPodSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomPodSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CustomPodSet{}, &CustomPodSetList{})
}
