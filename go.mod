module github.com/nokia/adcs-issuer

go 1.12

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20190918160344-1fbdaa4c8d90

require (
	github.com/Azure/go-ntlmssp v0.0.0-20180810175552-4a21cbd618b4
	github.com/go-logr/logr v0.1.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db // indirect
	github.com/jetstack/cert-manager v0.11.0
	github.com/onsi/ginkgo v1.10.2
	github.com/onsi/gomega v1.7.0
	k8s.io/api v0.0.0-20190918155943-95b840bb6a1f
	k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655
	//k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/klog v0.4.0
	//k8s.io/klog v0.4.0
	sigs.k8s.io/controller-runtime v0.4.0
)
