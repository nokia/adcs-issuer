/*

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

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	core "k8s.io/api/core/v1"

	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"

	api "github.com/chojnack/adcs-issuer/api/v1"
	"github.com/chojnack/adcs-issuer/issuers"
)

// AdcsRequestReconciler reconciles a AdcsRequest object
type AdcsRequestReconciler struct {
	client.Client
	Log                          logr.Logger
	IssuerFactory                issuers.IssuerFactory
	Recorder                     record.EventRecorder
	CertificateRequestController *CertificateRequestReconciler
}

// +kubebuilder:rbac:groups=adcs.certmanager.csf.nokia.com,resources=adcsrequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=adcs.certmanager.csf.nokia.com,resources=adcsrequests/status,verbs=get;update;patch

func (r *AdcsRequestReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("adcsrequest", req.NamespacedName)

	// your logic here
	log.Info("Processing request")

	// Fetch the AdcsRequest resource being reconciled
	ar := new(api.AdcsRequest)
	if err := r.Client.Get(ctx, req.NamespacedName, ar); err != nil {
		// We don't log error here as this is probably the 'NotFound'
		// case for deleted object.
		//
		// The Manager will log other errors.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// Find the issuer
	issuer, err := r.IssuerFactory.GetIssuer(ctx, ar.Spec.IssuerRef, ar.Namespace)
	if err != nil {
		log.WithValues("issuer", ar.Spec.IssuerRef).Error(err, "Couldn't get issuer")
		return ctrl.Result{}, err
	}

	cert, err := issuer.Issue(ctx, ar)
	if err != nil {
		// This is a local error.
		// We don't change the request status and just put it back on the queue
		// to re-try later.
		log.Error(err, fmt.Sprintf("Failed request will be re-tried in %v", issuer.RetryInterval))
		return ctrl.Result{Requeue: true, RequeueAfter: issuer.RetryInterval}, nil
	}

	// Get the original CertificateRequest to set result in
	cr, err := r.CertificateRequestController.GetCertificateRequest(ctx, req.NamespacedName)
	switch ar.Status.State {
	case api.Pending:
		// Check again later
		log.Info(fmt.Sprintf("Pending request will be re-tried in %v", issuer.StatusCheckInterval))
		r.setStatus(ctx, ar)
		return ctrl.Result{Requeue: true, RequeueAfter: issuer.StatusCheckInterval}, nil
	case api.Ready:
		cr.Status.Certificate = cert
		r.CertificateRequestController.SetStatus(ctx, &cr, cmmeta.ConditionTrue, cmapi.CertificateRequestReasonIssued, "ADCS request successfull")
	case api.Rejected:
		// This is a little hack for strange cert-manager behavior in case of failed request. Cert-manager automatically
		// re-tries such requests (re-created CertificateRequest object) what doesn't make sense in case of rejection.
		// We keep the Reason 'Pending' to prevent from re-trying while the actual status is in the Status Condition's Message field.
		// TODO: change it when cert-manager handles this better.
		r.CertificateRequestController.SetStatus(ctx, &cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonPending, "ADCS request rejected")
	case api.Errored:
		r.CertificateRequestController.SetStatus(ctx, &cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonFailed, "ADCS request errored")
	}
	r.setStatus(ctx, ar)

	return ctrl.Result{}, nil
}

func (r *AdcsRequestReconciler) setStatus(ctx context.Context, ar *api.AdcsRequest) error {

	// Fire an Event to additionally inform users of the change
	eventType := core.EventTypeNormal
	if ar.Status.State == api.Errored || ar.Status.State == api.Rejected {
		eventType = core.EventTypeWarning
	}
	r.Recorder.Event(ar, eventType, string(ar.Status.State), ar.Status.Reason)

	return r.Client.Status().Update(ctx, ar)
}

func (r *AdcsRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.AdcsRequest{}).
		Complete(r)
}
