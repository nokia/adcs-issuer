package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// ADCSRequest is a type to represent a certificate request within an ADCS server
// +k8s:openapi-gen=true
// +kubebuilder:printcolumn:name="Issuer",type="string",JSONPath=".spec.issuerRef.name",description="",priority=1
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=".status.reason",description="",priority=1
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=adcsrequests
type ADCSRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   ADCSRequestSpec   `json:"spec,omitempty"`
	Status ADCSRequestStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ADCSRequestList is a list of ADCSRequests
type ADCSRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ADCSRequest `json:"items"`
}

type ADCSRequestSpec struct {
	// Certificate signing request bytes in DER encoding.
	// This will be used when finalizing the request.
	// This field must be set on the request.
	CSR []byte `json:"csr"`

	// IssuerRef references a properly configured ADCS-type Issuer which should
	// be used to serve this ADCSRequest.
	// If the Issuer does not exist, processing will be retried.
	// If the Issuer is not an 'ADCS' Issuer, an error will be returned and the
	// ADCSRequest will be marked as failed.
	IssuerRef cmmeta.ObjectReference `json:"issuerRef"`
}

type ADCSRequestStatus struct {
	// ID of the Request assigned by the ADCS.
	// This will initially be empty when the resource is first created.
	// The ADCSRequest controller will populate this field when the Request is accepted by ADCS.
	// This field will be immutable after it is initially set.
	// +optional
	Id string `json:"id,omitempty"`

	// State contains the current state of this ADCSRequest resource.
	// States 'success' and 'expired' are 'final'
	// +optional
	State State `json:"state,omitempty"`

	// Reason optionally provides more information about a why the order is in
	// the current state.
	// +optional
	Reason string `json:"reason,omitempty"`
}

// State represents the state of an ADCSRequest.
// Clients utilising this type must also gracefully handle unknown
// values, as the contents of this enumeration may be added to over time.
// +kubebuilder:validation:Enum=valid;ready;pending;processing;invalid;expired;errored
type State string

const (
	// It is used to represent an unrecognised value.
	Unknown State = ""

	// If a request  is 'valid', the certificate has been issued by the ADCS server.
	// This is a final state.
	Valid State = "valid"

	// If a request is marked 'Pending', is's waiting for acceptance on the ADCS.
	// This is a transient state.
	Pending State = "pending"

	// Errored signifies that the ADCS request has errored for some reason.
	// This is a catch-all state, and is used for marking internal cert-manager
	// errors such as validation failures.
	// This is a final state.
	Errored State = "errored"
)
