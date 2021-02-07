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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DatabaseInstanceSpec defines the desired state of DatabaseInstance
type DatabaseInstanceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Postgres  *PostgresSpec  `json:"postgres,omitempty"`
	Clikhouse *ClikhouseSpec `json:"clickhouse,omitempty"`
}

type PostgresSpec struct {
	SqlParams
	AuthDB    string
	AuthDBRef *ParamRef
}

type ClikhouseSpec struct {
	SqlParams
}

type SqlParams struct {
	Username    string
	UsernameRef *ParamRef
	Password    string
	PasswordRef *ParamRef
	Host        string
	HostRef     *ParamRef
	Port        int
	PortRef     *ParamRef
}

type ParamRef struct {
	// Kind of the referent.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	Kind string `json:"kind,omitempty" protobuf:"bytes,1,opt,name=kind"`
	// Namespace of the referent.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/
	// +optional
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,2,opt,name=namespace"`
	// Name of the referent.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	// +optional
	Name string `json:"name,omitempty" protobuf:"bytes,3,opt,name=name"`
	// Data  key.
	// +optional
	Key string
}

// DatabaseInstanceStatus defines the observed state of DatabaseInstance
type DatabaseInstanceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Phase     Phase  `json:"phase,omitempty"`
	LastError string `json:"lastError,omitempty"`
}

type Phase string

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DatabaseInstance is the Schema for the databaseinstances API
type DatabaseInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseInstanceSpec   `json:"spec,omitempty"`
	Status DatabaseInstanceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DatabaseInstanceList contains a list of DatabaseInstance
type DatabaseInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DatabaseInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DatabaseInstance{}, &DatabaseInstanceList{})
}
