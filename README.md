# ADCS Issuer

ADCS Issuer is a [cert-manager's](https://github.com/jetstack/cert-manager) CertificateReuqest controller that uses MS Active Directory Certificate Service to sign certificates. 
It supports NTLM authentication.

## Description

### Requirements
ADCS Issuer has been tested with cert-manager v.0.11.0 and currently supports CertificateRequest CRD API version v1alpha2 only.

## Configuration
The ADCS service data can configured in AdcsIssuer or ClusterAdcsIssuer CRD objects.
```
kind: AdcsIssuer
metadata:
  name: test-adcs
  namespace: <namespace>
spec:
  caBundle: <base64-encoded-ca-certificate>
  credentialsRef:
    name: test-adcs-issuer-credentials
  statusCheckInterval: 6h
  retryInterval: 1h
  url: <adcs-certice-url>
```

ADCS Issuer creates AdcsRequest CRD objects 

## Installation

## Testing considerations

## Open issues
