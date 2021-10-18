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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ReceiverSpec defines the desired state of Receiver
type ReceiverSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Receiver. Edit receiver_types.go to remove/update
	Group          string         `json:"group,omitempty"`
	DefaultMethond bool           `json:"defaultMethond,omitempty"`
	WebhookConfig  *WebhookConfig `json:"webhookConfig,omitempty"`
	DogConfig      *DogConfig     `json:"DogConfig,omitempty"`
	YachConfig     *YachConfig    `json:"yachConfig,omitempty"`
}

// ReceiverStatus defines the observed state of Receiver
type ReceiverStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// Receiver is the Schema for the receivers API
type Receiver struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReceiverSpec   `json:"spec,omitempty"`
	Status ReceiverStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ReceiverList contains a list of Receiver
type ReceiverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Receiver `json:"items"`
}

// WebhookConfig configures notifications via a generic receiver supporting the webhook payload.
// See https://prometheus.io/docs/alerting/latest/configuration/#webhook_config
type WebhookConfig struct {
	// Whether or not to notify about resolved alerts.
	// +optional
	SendResolved *bool `json:"sendResolved,omitempty"`
	// The URL to send HTTP POST requests to. `urlSecret` takes precedence over
	// `url`. One of `urlSecret` and `url` should be defined.
	// +optional
	URL *string `json:"url,omitempty"`
	// The secret's key that contains the webhook URL to send HTTP requests to.
	// `urlSecret` takes precedence over `url`. One of `urlSecret` and `url`
	// should be defined.
	// The secret needs to be in the same namespace as the AlertmanagerConfig
	// object and accessible by the Prometheus Operator.
	// +optional
	URLSecret *v1.SecretKeySelector `json:"urlSecret,omitempty"`
	// HTTP client configuration.
	// +optional
	HTTPConfig *HTTPConfig `json:"httpConfig,omitempty"`
	// Maximum number of alerts to be sent per webhook message. When 0, all alerts are included.
	// +optional
	// +kubebuilder:validation:Minimum=0
	MaxAlerts int32 `json:"maxAlerts,omitempty"`
}

// BasicAuth allow an endpoint to authenticate over basic authentication
// More info: https://prometheus.io/docs/operating/configuration/#endpoints
// +k8s:openapi-gen=true
type BasicAuth struct {
	// The secret in the service monitor namespace that contains the username
	// for authentication.
	Username v1.SecretKeySelector `json:"username,omitempty"`
	// The secret in the service monitor namespace that contains the password
	// for authentication.
	Password v1.SecretKeySelector `json:"password,omitempty"`
}

// SafeTLSConfig specifies safe TLS configuration parameters.
// +k8s:openapi-gen=true
type SafeTLSConfig struct {
	// Struct containing the CA cert to use for the targets.
	CA SecretOrConfigMap `json:"ca,omitempty"`
	// Struct containing the client cert file for the targets.
	Cert SecretOrConfigMap `json:"cert,omitempty"`
	// Secret containing the client key file for the targets.
	KeySecret *v1.SecretKeySelector `json:"keySecret,omitempty"`
	// Used to verify the hostname for the targets.
	ServerName string `json:"serverName,omitempty"`
	// Disable target certificate validation.
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`
}

// SecretOrConfigMap allows to specify data as a Secret or ConfigMap. Fields are mutually exclusive.
type SecretOrConfigMap struct {
	// Secret containing data to use for the targets.
	Secret *v1.SecretKeySelector `json:"secret,omitempty"`
	// ConfigMap containing data to use for the targets.
	ConfigMap *v1.ConfigMapKeySelector `json:"configMap,omitempty"`
}

// HTTPConfig defines a client HTTP configuration.
// See https://prometheus.io/docs/alerting/latest/configuration/#http_config
type HTTPConfig struct {
	// BasicAuth for the client.
	// +optional
	BasicAuth *BasicAuth `json:"basicAuth,omitempty"`
	// The secret's key that contains the bearer token to be used by the client
	// for authentication.
	// The secret needs to be in the same namespace as the AlertmanagerConfig
	// object and accessible by the Prometheus Operator.
	// +optional
	BearerTokenSecret *v1.SecretKeySelector `json:"bearerTokenSecret,omitempty"`
	// TLS configuration for the client.
	// +optional
	TLSConfig *SafeTLSConfig `json:"tlsConfig,omitempty"`
	// Optional proxy URL.
	// +optional
	ProxyURL string `json:"proxyURL,omitempty"`
}

type DogConfig struct {
	MaxAlerts int32 `json:"maxAlerts,omitempty"`
	TaskId    int32 `json:"taskId,omitempty"`
}

type YachConfig struct {
	AccessToken string `json:"accessToken"`
	Secret      string `json:"secret,omitempty"`
	Keyword     string `json:"keyword,omitempty"`
}
