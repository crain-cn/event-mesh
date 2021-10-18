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

package v1

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// EventRouteSpec defines the desired state of EventRoute
type EventRouteSpec struct {
	Route *Route `json:"route"`
	//Receivers []*CrossVersionObjectReference `json:"receivers"`
}

// Route defines a node in the routing tree.
type Route struct {
	// Name of the receiver for this route. If not empty, it should be listed in
	// the `receivers` field.
	// +optional
	Receiver string `json:"receiver" example:"Receiver name"`
	// List of labels to group by.
	// +optional
	GroupBy []string `json:"groupBy,omitempty"`
	// How long to wait before sending the initial notification. Must match the
	// regular expression `[0-9]+(ms|s|m|h)` (milliseconds seconds minutes
	// hours).
	// +optional
	GroupWait string `json:"groupWait,omitempty"`
	// How long to wait before sending an updated notification. Must match the
	// regular expression `[0-9]+(ms|s|m|h)` (milliseconds seconds minutes
	// hours).
	// +optional
	GroupInterval string `json:"groupInterval,omitempty"`
	// How long to wait before repeating the last notification. Must match the
	// regular expression `[0-9]+(ms|s|m|h)` (milliseconds seconds minutes
	// hours).
	// +optional
	RepeatInterval string `json:"repeatInterval,omitempty"`
	// List of matchers that the alert’s labels should match. For the first
	// level route, the operator removes any existing equality and regexp
	// matcher on the `namespace` label and adds a `namespace: <object
	// namespace>` matcher.
	// +optional
	Matchers []Matcher `json:"matchers,omitempty"`
	// Boolean indicating whether an alert should continue matching subsequent
	// sibling nodes. It will always be overridden to true for the first-level
	// route by the Prometheus operator.
	// +optional
	Continue bool `json:"continue,omitempty"`
	// Child routes.
	Routes []apiextensionsv1.JSON `json:"routes,omitempty"`
	// Note: this comment applies to the field definition above but appears
	// below otherwise it gets included in the generated manifest.
	// CRD schema doesn't support self referential types for now (see
	// https://github.com/kubernetes/kubernetes/issues/62872). We have to use
	// an alternative type to circumvent the limitation. The downside is that
	// the Kube API can't validate the data beyond the fact that it is a valid
	// JSON representation.
}

// Matcher defines how to match on alert's labels.
type Matcher struct {
	// Label to match.
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
	// Label value to match.
	// +optional
	Value string `json:"value"`
	// Whether to match on equality (false) or regular-expression (true).
	// +optional
	Regex bool `json:"regex,omitempty"`
}

// EventRouteStatus defines the observed state of EventRoute
type EventRouteStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// EventRoute is the Schema for the eventroutes API
type EventRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EventRouteSpec   `json:"spec,omitempty"`
	Status EventRouteStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// EventRouteList contains a list of EventRoute
type EventRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EventRoute `json:"items"`
}

// CrossVersionObjectReference contains enough information to let you identify the referred resource.
type CrossVersionObjectReference struct {
	// Kind of the referent; More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds"
	Kind string `json:"kind" protobuf:"bytes,1,opt,name=kind"`
	// Name of the referent; More info: http://kubernetes.io/docs/user-guide/identifiers#names
	Name string `json:"name" protobuf:"bytes,2,opt,name=name"`
	// API version of the referent
	// +optional
	APIVersion string `json:"apiVersion,omitempty" protobuf:"bytes,3,opt,name=apiVersion"`
}
