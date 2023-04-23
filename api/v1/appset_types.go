/*
Copyright 2023 zhengyansheng.

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

// AppSetSpec defines the desired state of AppSet
type AppSetSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of AppSet. Edit appset_types.go to remove/update
	Name         string `json:"name,omitempty"` // monkey
	Namespace    string `json:"namespace"`      // namespace
	Replicas     int32  `json:"replicas"`       // 2
	Image        string `json:"image"`          // nginx:latest
	ExposePort   int32  `json:"expose_port"`    // 80
	ExposeDomain string `json:"expose_domain"`  // www.monkey.com
}

// AppSetStatus defines the observed state of AppSet
type AppSetStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Replicas *int32 `json:"replicas,omitempty"` // 2
	Ready    bool   `json:"ready"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AppSet is the Schema for the appsets API
type AppSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppSetSpec   `json:"spec,omitempty"`
	Status AppSetStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AppSetList contains a list of AppSet
type AppSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AppSet{}, &AppSetList{})
}
