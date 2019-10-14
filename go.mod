module github.com/chojnack/adcs-issuer

go 1.12

replace (
	k8s.io/api => k8s.io/api v0.0.0-20190718183219-b59d8169aab5
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190612205821-1799e75a0719

	k8s.io/client-go => k8s.io/client-go v0.0.0-20190718183610-8e956561bbf5
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.3.0
)

require (
	github.com/go-logr/logr v0.1.0
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db // indirect
	github.com/jetstack/cert-manager v0.11.0
	github.com/onsi/ginkgo v1.10.2
	github.com/onsi/gomega v1.7.0
	k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.2.0-beta.4
)
