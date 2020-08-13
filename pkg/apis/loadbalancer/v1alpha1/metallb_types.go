package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type BGPPeer struct {
	// Peer address
	Address string `json:"peer-address"`

	// Peer ASN
	PeerASN string `json:"peer-asn"`

	// My ASN
	ASN     string `json:"my-asn"`
}

type AddressPool struct {
	// Name of the pool
	Name string `json:"name"`

	// The protocol to use for advertising VIPs. Either layer2 or bgp.
	Protocol string `json:"protocol"`

	// A range of VIPs that metallb can use for loadbalancing. Can be either a
	// full CIDR or a range like 1.1.1.1-1.1.1.10.
	Addresses []string `json:"addresses"`

	// TODO(bnemec): bgp-advertisements
}

// MetalLBSpec defines the desired state of MetalLB
type MetalLBSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// Address pools
	AddressPools []AddressPool `json:"address-pools"`

	// Peers for BGP mode
	Peers []BGPPeer `json:"peers,omitempty"`

	// TODO(bnemec): bgp-communities
}

// MetalLBStatus defines the observed state of MetalLB
type MetalLBStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MetalLB is the Schema for the metallbs API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=metallbs,scope=Namespaced
type MetalLB struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MetalLBSpec   `json:"spec,omitempty"`
	Status MetalLBStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MetalLBList contains a list of MetalLB
type MetalLBList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MetalLB `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MetalLB{}, &MetalLBList{})
}
