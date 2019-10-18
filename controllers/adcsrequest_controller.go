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
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	adcsv1 "github.com/chojnack/adcs-issuer/api/v1"
	"github.com/chojnack/adcs-issuer/issuers"
)

// AdcsRequestReconciler reconciles a AdcsRequest object
type AdcsRequestReconciler struct {
	client.Client
	Log           logr.Logger
	IssuerFactory issuers.IssuerFactory
	Recorder      record.EventRecorder
}

// +kubebuilder:rbac:groups=adcs.certmanager.csf.nokia.com,resources=adcsrequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=adcs.certmanager.csf.nokia.com,resources=adcsrequests/status,verbs=get;update;patch

func (r *AdcsRequestReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("adcsrequest", req.NamespacedName)

	// your logic here
	log.Info("Processing request")

	// Fetch the AdcsRequest resource being reconciled
	ar := new(adcsv1.AdcsRequest)
	if err := r.Client.Get(ctx, req.NamespacedName, ar); err != nil {
		// We don't log error here as this is probably the 'NotFound'
		// case for deleted object.
		//
		// The Manager will log other errors.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	issuer, err := r.IssuerFactory.GetIssuer(ctx, ar.Spec.IssuerRef, ar.Namespace)
	if err != nil {
		log.WithValues("issuer", ar.Spec.IssuerRef).Error(err, "Couldn't get issuer")
		return ctrl.Result{}, err
	}

	log.Info(fmt.Sprintf("Using issuer user=%s password=%s url=%s", issuer.Username, issuer.Password, issuer.Url))

	return ctrl.Result{}, nil
}

func (r *AdcsRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&adcsv1.AdcsRequest{}).
		Complete(r)
}
