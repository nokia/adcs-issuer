package issuers

import (
	"context"
	"fmt"
	"time"

	//cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	//cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chojnack/adcs-issuer/adcs"
	api "github.com/chojnack/adcs-issuer/api/v1"
)

const (
	adcsCertTemplate = "BasicSSLWebServer"
)

type Issuer struct {
	client.Client
	certServ            adcs.AdcsCertsrv
	RetryInterval       time.Duration
	StatusCheckInterval time.Duration
}

func (i *Issuer) Issue(ctx context.Context, ar *api.AdcsRequest) ([]byte, error) {
	var adcsResponseStatus adcs.AdcsResponseStatus
	var desc string
	var id string
	var err error
	if ar.Status.State != api.Unknown {
		// Of all the statuses only Pending requires processing.
		// All others are final
		if ar.Status.State == api.Pending {
			// Check the status of the reqeust on the ADCS
			if ar.Status.Id == "" {
				return nil, fmt.Errorf("ADCS ID not set.")
			}
			adcsResponseStatus, desc, id, err = i.certServ.GetExistingCertificate(ar.Status.Id)
		} else {
			// Nothing to do
			return nil, nil
		}
	} else {
		// New request
		adcsResponseStatus, desc, id, err = i.certServ.RequestCertificate(string(ar.Spec.CSRPEM), adcsCertTemplate)
	}
	if err != nil {
		// This is a local error
		return nil, err
	}

	var cert []byte
	switch adcsResponseStatus {
	case adcs.Pending:
		// It must be checked again later
		ar.Status.State = api.Pending
		ar.Status.Id = id
		ar.Status.Reason = desc
	case adcs.Ready:
		// Certificate obtained successfully
		ar.Status.State = api.Ready
		ar.Status.Id = id
		ar.Status.Reason = ""
		cert = []byte(desc)
	case adcs.Rejected:
		// Certificate request rejected by ADCS
		ar.Status.State = api.Rejected
		ar.Status.Id = id
		ar.Status.Reason = desc
	case adcs.Errored:
		// Unknown problem occured on ADCS
		ar.Status.State = api.Errored
		ar.Status.Id = id
		ar.Status.Reason = desc
	}

	return cert, nil

}
