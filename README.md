# ADCS Issuer



ADCS Issuer is a [cert-manager's](https://github.com/jetstack/cert-manager) CertificateRequest controller that uses MS Active Directory Certificate Service to sign certificates 
(see [this design document](https://github.com/jetstack/cert-manager/blob/master/design/20190708.certificate-request-crd.md) for details on CertificateRequest CRD). 

ADCS provides HTTP GUI that can be normally used to request new certificates or see status of existing requests. This implementation is simply a HTTP client that interacts with the
ADCS server sending appropriately prepared HTTP requests and interpretting the server's HTTP responses (the approach inspired by [this Python ADCS client](https://github.com/magnuswatn/certsrv)).
It supports NTLM authentication.

## Description

### Requirements
ADCS Issuer has been tested with cert-manager v.0.11.0 and currently supports CertificateRequest CRD API version v1alpha2 only.

## Configuration and usage

### Issuers
The ADCS service data can configured in `AdcsIssuer` or `ClusterAdcsIssuer` CRD objects e.g.:
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
If cluster level issuer configuration is needed then ClusterAdcsUssuer can be defined like this:
```
kind: ClusterAdcsIssuer
metadata:
  name: test-adcs
spec:
  caBundle: <base64-encoded-ca-certificate>
  credentialsRef:
    name: test-adcs-issuer-credentials
  statusCheckInterval: 6h
  retryInterval: 1h
  url: <adcs-certice-url>
```
The secret used by the `ClusterAdcsIssuer` must be defined in the namespace where controller's pod is running.

### Requesting certificates

To request a certificate with `AdcsIssuer` the standard `certificate.cert-manager.io` object needs to be created. The `issuerRef` must be set to point to `AdcsIssuer` or `ClusterAdcsIssuer` object
from group `adcs.certmanager.csf.nokie.com` e.g.:
```
apiVersion: cert-manager.io/v1alpha2
kind: Certificate
metadata:
  annotations:
  name: adcs-cert
  namespace: <namespace>
spec:
  commonName: example.com
  dnsNames:
  - service1.example.com
  - service2.example.com
  issuerRef:
    group: adcs.certmanager.csf.nokia.com
    kind: AdcsIssuer
    name: test-adcs
  organization:
  - Your organization
  secretName: adcs-cert
```
Cert-manager is responsible for creating the `Secret` with a key and `CertificateRequest` with proper CSR data.


ADCS Issuer creates `AdcsRequest` CRD object that keep actual state of the processing. Its name is always the same as the corresponding `CertificateRequest` object (there is strict one-to-one mapping).
The `AdcsRequest` object stores the ID of request assigned by the ADCS server as wall as the current status which can be one of:
* **Pending** - the request has been sent to ADCS and is waiting for acceptance (status will be checked periodically),
* **Ready** - the request has been successfully processed and the certificate is ready and stored in secret defined in the original `Certificate` object,
* **Rejected** - the request was rejected by ADCS and will be re-tried unless the `Certificate` is updated,
* **Errored**  - unrecoverable problem occured.

```
apiVersion: adcs.certmanager.csf.nokia.com/v1
kind: AdcsRequest
metadata:
  name: adcs-cert-3831834799
  namespace: c1
  ownerReferences:
  - apiVersion: cert-manager.io/v1alpha2
    blockOwnerDeletion: true
    controller: true
    kind: CertificateRequest
    name: adcs-cert-3831834799
    uid: f5cf630d-f4cf-11e9-95eb-fa163e038ef8
  uid: f5d22b47-f4cf-11e9-95eb-fa163e038ef8
spec:
  csr: <base64-encoded-csr>
  issuerRef:
    group: adcs.certmanager.csf.nokia.com
    kind: AdcsIssuer
    name: test-adcs
status:
  id: "18"
  state: ready
```

#### Auto-request certificate from ingress
Add the following to an `Ingress` for cert-manager to auto-generate a
`Certificate` using `Ingress` information with ingress-shim
```
metadata:
  name: test-ingress
    annotations:
        cert-manager.io/issuer: "adcs-issuer" #use specific name of issuer
        cert-manager.io/issuer-kind: "AdcsIssuer" #or AdcsClusterIssuer
        cert-manager.io/issuer-group: "adcs.certmanager.csf.nokia.com"
```
in addition to
```
spec:
  tls:
    - hosts:
        - test-host.com
            secretName: ingress-secret # secret cert-manager stores certificate in
```

## Installation

This controller is implemented using [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder). Automatically generated Makefile contains targets needed for build and installation. 
Generated CRD manifests are stored in `config/crd`. RBAC roles and bindings can be found in config/rbac. There's also a Make target to build controller's Docker image and
store it in local docker repo (Docker must be installed).

## Testing considerations

### ADCS Simulator
The test/adcs-sim directory contains a simple ADCS simulator that can be used for basic tests (run `make sim-install` to build it and install in /usr/local directory tree). The simulator can be started on the host and work ad ADCS server that will sign certificates using provided self-signed certificate and key (`root.pem` and `root.key` files). If needed the certificate can be replaced with any other available.

The simulator accepts directives to control its behavior. The directives are set as additional domain names in the certificate request:
* **delay.<time>.sim**  where <time> is e.g. 10m, 15h etc - the certificate will be issued after the specified time
* **reject.sim** - the certificate will be rejected
* **unauthorized.sim** - the certificate request will be rejected because of authorization problems (to simulate invalid user permissions)

More then one directive can be used at a time. e.g. to simulate rejecting the certificate after 10 minutes add the following domain names:

```
- delay.10m.sim
- reject.sim
```

## Open issues
 
* Cert-manger limits the identity of the requestor to Organization and CommonName. Full X509 Distinguished Name support is needed. See: [Full X509 Distinguished Name support](https://github.com/jetstack/cert-manager/issues/2288)
* When request is rejected by ADCS because of invalid data then there's a problem to indicate in CertificateReuqest that it should not be re-tried. See: [Problem with automatic retry of failed requests](https://github.com/jetstack/cert-manager/issues/2289)

## ToDos

* Webhook
* Helm chart
* ...

