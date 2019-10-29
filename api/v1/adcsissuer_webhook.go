package v1

import (
	"regexp"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	//validationutils "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/jetstack/cert-manager/pkg/util/pki"
)

var log = logf.Log.WithName("adcsissuer-resource")

func (r *AdcsIssuer) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-adcs-certmanager-csf-nokia-com-v1-adcsissuer,mutating=true,failurePolicy=fail,groups=adcs.certmanager.csf.nokia.com,resources=adcsissuer,verbs=create;update,versions=v1,name=adcsissuer-mutation.adcs.certmanager.csf.nokia.com

var _ webhook.Defaulter = &AdcsIssuer{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *AdcsIssuer) Default() {
	log.Info("default", "name", r.Name)

	if r.Spec.StatusCheckInterval == "" {
		r.Spec.StatusCheckInterval = "6h"
	}
	if r.Spec.RetryInterval == "" {
		r.Spec.RetryInterval = "1h"
	}
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-adcs-certmanager-csf-nokia-com-v1-adcsissuer,mutating=false,failurePolicy=fail,groups=adcs.certmanager.csf.nokia.com,resources=adcsissuer,versions=v1,name=adcsissuer-validation.adcs.certmanager.csf.nokia.com

var _ webhook.Validator = &AdcsIssuer{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *AdcsIssuer) ValidateCreate() error {
	log.Info("validate create", "name", r.Name)

	return r.validateAdcsIssuer()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *AdcsIssuer) ValidateUpdate(old runtime.Object) error {
	log.Info("validate update", "name", r.Name)

	return r.validateAdcsIssuer()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *AdcsIssuer) ValidateDelete() error {
	log.Info("validate delete", "name", r.Name)

	return nil
}

func (r *AdcsIssuer) validateAdcsIssuer() error {
	var allErrs field.ErrorList

	// Validate RetryInterval
	_, err := time.ParseDuration(r.Spec.RetryInterval)
	if err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("retryInterval"), r.Spec.RetryInterval, err.Error()))
	}

	// Validate Status Check Interval
	_, err = time.ParseDuration(r.Spec.StatusCheckInterval)
	if err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("statusCheckInterval"), r.Spec.StatusCheckInterval, err.Error()))
	}

	// Validate URL. Must be valide http or https URL
	re := regexp.MustCompile(`(http|https):\/\/([\w\-_]+(?:(?:\.[\w\-_]+)+))([\w\-\.,@?^=%&amp;:/~\+#]*[\w\-\@?^=%&amp;/~\+#])?`)
	if !re.MatchString(r.Spec.URL) {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("url"), r.Spec.URL, "Invalid URL format. Must be valid 'http://' or 'https://' URL."))
	}

	// Validate CA Bundle. Must be a valid certificate PEM.
	_, err = pki.DecodeX509CertificateBytes(r.Spec.CABundle)
	if err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("caBundle"), r.Spec.CABundle, err.Error()))
	}

	// TODO: Validate credentials secret name?

	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(
		schema.GroupKind{Group: "adcs.certmanager.csf.nokia.com", Kind: "AdcsIssuer"},
		r.Name, allErrs)

}
