package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TimSecretConfigSpec defines the Vault configuration
type TimSecretConfigSpec struct {
	// VaultURL is the Vault server URL
	VaultURL string `json:"vaultURL"`

	// VaultToken is the authentication token for Vault
	VaultToken string `json:"vaultToken"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced

// TimSecretConfig is the Schema for centralized Vault configuration
type TimSecretConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec TimSecretConfigSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// TimSecretConfigList contains a list of TimSecretConfig
type TimSecretConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TimSecretConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TimSecretConfig{}, &TimSecretConfigList{})
}
