# ADCS Issuer

ADCS Issuer is a [cert-manager's](https://github.com/jetstack/cert-manager) CertificateReuqest controller that uses MS Active Directory Certificate Service to sign certificates. 
It supports NTLM authentication.

ADCS provides HTTP GUI that can be normally used to request new certificates or see status of existing requests. This implementation is simply a HTTP client that interacts with the
ADCS server sending appropriately prepared HTTP requests and interpretting the server's HTTP responses (the approach inspired by [this Python ADCS client](https://github.com/magnuswatn/certsrv)).

## Description

### Requirements
ADCS Issuer has been tested with cert-manager v.0.11.0 and currently supports CertificateRequest CRD API version v1alpha2 only.

## Configuration
The ADCS service data can configured in AdcsIssuer or ClusterAdcsIssuer CRD objects e.g.:
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

The `caBundle` parameter is BASE64-encoded CA certificate which is used by the ADCS server itself, which may not be the same certificate that will be used to sign your request.

The `statusCheckInterval` indicates how often the status of the request should be tested. Typically, it can take a few hours or even days before the certificate is issued.

The `retryInterval` says how long to wait before retrying requests that errored.

The `credentialsRef.name` is name of a secret that stores user credentials used for NTLM authentication. The secret must be `Opaque` and contain `password` and `username` fields only e.g.:
```
apiVersion: v1
data:
  password: cGFzc3dvcmQ=
  username: dXNlcm5hbWU=
kind: Secret
metadata:
  name: test-adcs-issuer-credentials
  namespace: <namespace>
type: Opaque
```



ADCS Issuer creates AdcsRequest CRD objects 

## Installation

## Testing considerations

## Open issues
