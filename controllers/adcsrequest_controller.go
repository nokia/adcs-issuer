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

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	adcsv1 "github.com/chojnack/adcs-issuer/api/v1"
)

// AdcsRequestReconciler reconciles a AdcsRequest object
type AdcsRequestReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=adcs.certmanager.csf.nokia.com,resources=adcsrequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=adcs.certmanager.csf.nokia.com,resources=adcsrequests/status,verbs=get;update;patch

func (r *AdcsRequestReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("adcsrequest", req.NamespacedName)

	// your logic here

	return ctrl.Result{}, nil
}

func (r *AdcsRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&adcsv1.AdcsRequest{}).
		Complete(r)
}
