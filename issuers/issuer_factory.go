package issuers

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"

	//apimachtypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"

	adcsv1 "github.com/chojnack/adcs-issuer/api/v1"
)

type IssuerFactory struct {
	client.Client
}

func (f *IssuerFactory) GetIssuer(ctx context.Context, ref cmmeta.ObjectReference, namespace string) (*Issuer, error) {
	key := client.ObjectKey{Namespace: namespace, Name: ref.Name}

	switch strings.ToLower(ref.Kind) {
	case "adcsissuer":
		return f.getAdcsIssuer(ctx, key)
	case "clusteradcsissuer":
		return f.getClusterAdcsIssuer(ctx, key)
	}
	return nil, fmt.Errorf("Unsupported issuer kind %s.", ref.Kind)
}

// Get AdcsIssuer object from K8s
func (f *IssuerFactory) getAdcsIssuer(ctx context.Context, key client.ObjectKey) (*Issuer, error) {
	iss := new(adcsv1.AdcsIssuer)
	if err := f.Client.Get(ctx, key, iss); err != nil {
		return nil, err
	}
	// TODO: add checking issuer status

	username, password, err := f.getUserPassword(ctx, iss)
	if err != nil {
		return nil, err
	}
	return &Issuer{
		f.Client,
		username,
		password,
		iss.Spec.URL,
	}, nil
}

// Get ClusterAdcsIssuer object from K8s
func (f *IssuerFactory) getClusterAdcsIssuer(ctx context.Context, key client.ObjectKey) (*Issuer, error) {

	return nil, nil
}

func (f *IssuerFactory) getUserPassword(ctx context.Context, iss *adcsv1.AdcsIssuer) (string, string, error) {
	secret := new(corev1.Secret)
	if err := f.Client.Get(ctx, client.ObjectKey{Namespace: iss.Namespace, Name: iss.Spec.CredentialsRef.Name}, secret); err != nil {
		return "", "", err
	}
	if _, ok := secret.Data["username"]; !ok {
		return "", "", fmt.Errorf("User name not set in secret")
	}
	if _, ok := secret.Data["password"]; !ok {
		return "", "", fmt.Errorf("Password not set in secret")
	}
	return string(secret.Data["username"]), string(secret.Data["password"]), nil
}
