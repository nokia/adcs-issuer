module github.com/nokia/adcs-issuer

go 1.16

require (
	github.com/Azure/go-ntlmssp v0.0.0-20200615164410-66371956d46c
	github.com/go-logr/logr v0.3.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/jetstack/cert-manager v1.3.1
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	k8s.io/api v0.20.2
	k8s.io/apiextensions-apiserver v0.20.2 // indirect
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	k8s.io/klog v1.0.0
	k8s.io/utils v0.0.0-20210111153108-fddb29f9d009
	sigs.k8s.io/controller-runtime v0.8.3
)
