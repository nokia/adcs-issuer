package issuers

import (
	//"context"
	//api "github.com/chojnack/adcs-issuer/api/v1"
	//cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	//cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Issuer struct {
	client.Client
	Username string
	Password string
	Url      string
}

/*
func GetIssuer(issType string, issName string) (*Issuer, error) {
	return nil, nil
}
*/

/*
func GetIssuer(ref cmmeta.ObjectReference) *Issuer {
	iss, exists := cache[ref]
	if exists {
		return iss
	}
	// Get from K8s

	return nil
}
*/

// Issuers and cluster issuers are cached together.
func getCacheKey(issType string, issName string) string {
	return issType + "/" + issName
}

// Register issuer detected by the controller.
// The registration includes:
// - validating ADCS credentials
// - checking ADCS connection
// - updating issuers cache
func RegisterIssuer(issType string, issName string) error {
	return nil
}
