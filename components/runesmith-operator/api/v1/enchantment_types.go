/*
Copyright 2025.

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
	"github.com/fukaraca/runesmith/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type EnchantmentSpecArtifact struct {
	ID int `json:"id"`

	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Common;Rare;Epic;Legendary
	// +kubebuilder:validation:Type=string
	Tier shared.Tier `json:"tier"`

	Requirements []EnchantmentSpecArtifactRequirement `json:"requirements"`

	Priority int `json:"priority"`
}

type EnchantmentSpecArtifactRequirement struct {
	// +kubebuilder:validation:Enum=fire;frost;arcane
	// +kubebuilder:validation:Type=string
	EnergyType shared.Elemental `json:"energyType"`

	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Type=string
	ResourceName shared.Resource `json:"resourceName"`

	// +kubebuilder:validation:Minimum=1
	Limit int `json:"limit"`
}

type EnchantmentRetentionPolicy struct {
	// +kubebuilder:validation:Minimum=5
	TTLSecondsAfterFinished *int `json:"ttlSecondsAfterFinished,omitempty"`
}

// EnchantmentSpec defines the desired state of Enchantment
type EnchantmentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	Retention EnchantmentRetentionPolicy `json:"retention,omitempty"`

	// +kubebuilder:validation:Minimum=1
	OrderID int `json:"orderId"`

	Artifact EnchantmentSpecArtifact `json:"artifact"`

	// +kubebuilder:validation:Minimum=1
	Cost int `json:"cost"`

	// +kubebuilder:default=true
	SelfReport *bool `json:"selfReport,omitempty"`
}

// EnchantmentStatus defines the observed state of Enchantment.
type EnchantmentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Enum=Scheduled;Enchanting;Failed;Completed;Requeued
	// +kubebuilder:validation:Type=string
	Phase shared.EnchantmentPhase `json:"phase,omitempty"`

	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
	ExpiresAt      *metav1.Time `json:"expiresAt,omitempty"`

	SucceededJobs int    `json:"succeededJobs"`
	FailedJobs    int    `json:"failedJobs"`
	ActiveJobs    int    `json:"activeJobs"`
	Progress      string `json:"progress,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Jobs",type=string,JSONPath=`.status.progress`
// +kubebuilder:printcolumn:name="Success",type=integer,priority=1,JSONPath=`.status.succeededJobs`
// +kubebuilder:printcolumn:name="Fail",type=integer,priority=1,JSONPath=`.status.failedJobs`
// +kubebuilder:printcolumn:name="Active",type=integer,priority=1,JSONPath=`.status.activeJobs`
// +kubebuilder:resource:shortName=ench

// Enchantment is the Schema for the enchantments API
type Enchantment struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of Enchantment
	// +required
	Spec EnchantmentSpec `json:"spec"`

	// status defines the observed state of Enchantment
	// +optional
	Status EnchantmentStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// EnchantmentList contains a list of Enchantment
type EnchantmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Enchantment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Enchantment{}, &EnchantmentList{})
}
