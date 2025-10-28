package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TimSecretSpec defines the desired state of TimSecret
type TimSecretSpec struct {
	// VaultConfig is the name of the TimSecretConfig to use
	// +optional
	VaultConfig string `json:"vaultConfig,omitempty"`

	// VaultConfigNamespace is the namespace of the TimSecretConfig
	// If not specified, uses the TimSecret's namespace
	// +optional
	VaultConfigNamespace string `json:"vaultConfigNamespace,omitempty"`

	// VaultURL is the Vault server URL (direct value, overrides VaultConfig)
	// +optional
	VaultURL string `json:"vaultURL,omitempty"`

	// VaultToken is the authentication token for Vault (direct value, overrides VaultConfig)
	// +optional
	VaultToken string `json:"vaultToken,omitempty"`

	// VaultPath is the path in Vault where secrets are stored
	VaultPath string `json:"vaultPath"`

	// SecretName is the name of the Kubernetes Secret to create
	SecretName string `json:"secretName"`

	// DeploymentName is the name of the Deployment to restart when secret changes
	// +optional
	DeploymentName string `json:"deploymentName,omitempty"`

	// Namespace is the namespace where the secret and deployment are located
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// TimSecretStatus defines the observed state of TimSecret
type TimSecretStatus struct {
	// LastSyncTime is the last time the secret was synced from Vault
	// +optional
	LastSyncTime *metav1.Time `json:"lastSyncTime,omitempty"`

	// SecretHash is the hash of the secret data for change detection
	// +optional
	SecretHash string `json:"secretHash,omitempty"`

	// Conditions represent the latest available observations of the TimSecret's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced

// TimSecret is the Schema for the timsecrets API
type TimSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TimSecretSpec   `json:"spec,omitempty"`
	Status TimSecretStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TimSecretList contains a list of TimSecret
type TimSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TimSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TimSecret{}, &TimSecretList{})
}
