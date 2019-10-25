package certserv

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	mrand "math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/jetstack/cert-manager/pkg/util/pki"
)

type Certserv struct {
	// TODO Initialize
	currentID uint64
	certs     []Cert
	caCert    *x509.Certificate
	caKey     *rsa.PrivateKey
}

var (
	caWorkDir  = *flag.String("workdir", "/usr/local/adcs-sim", "ADCS simulator working directory")
	caCertFile = caWorkDir + "/ca/root.pem"
	caKeyFile  = caWorkDir + "/ca/root.key"
	caDir      = caWorkDir + "/ca"

	tmplCertnewCer   = caWorkDir + "/templates/certnew.cer.tmpl"
	tmplCertCaRc     = caWorkDir + "/templates/certcarc.asp.tmpl"
	tmplCertFnsh     = caWorkDir + "/templates/certfnsh.asp.tmpl"
	tmplUnauthorized = caWorkDir + "/templates/unauth.tmpl"
)

type SimOrders struct {
	reject       bool
	delay        time.Duration
	unauthorized bool
}

func NewCertserv() (*Certserv, error) {
	cs := &Certserv{
		0,
		nil,
		nil,
		nil,
	}
	err := cs.initRootCert()
	if err != nil {
		return nil, fmt.Errorf("Error: %s", err.Error())
	}
	return cs, nil
}

func (c *Certserv) HandleCertnewCer(w http.ResponseWriter, req *http.Request) {
	tmpl, _ := template.ParseFiles(tmplCertnewCer)

	type Resp struct {
		DispositionMessage string
		LastStatus         string
	}
	err := req.ParseForm()
	if err != nil {
		respondError(w, "Cannot parse parameters")
		return
	}

	reqId := req.Form["ReqID"]
	if reqId == nil {
		respondError(w, "Missing ReqID")
		return
	}
	if reqId[0] == "CACert" {
		file, err := ioutil.ReadFile(caCertFile)
		if err != nil {
			respondError(w, "Cannot find root CA cert.")
			res := Resp{"Cannot find root CA cert.", "Error"}
			tmpl.Execute(w, res)
			return
		}
		w.Header().Add("Content-Type", "application/pkix-cert")
		fmt.Fprintf(w, "%s", file)
		return
	}
	certFileName := fmt.Sprintf("%s/%s.pem", caDir, reqId[0])
	csrFileName := fmt.Sprintf("%s/%s.csr", caDir, reqId[0])

	file, err := ioutil.ReadFile(certFileName)
	if err == nil {
		// Certificate file exists, so let's send it back
		w.Header().Add("Content-Type", "application/pkix-cert")
		fmt.Fprintf(w, "%s", file)
		return
	} else if !os.IsNotExist(err) {
		// Error other than 'file doesn't exists' occured
		msg := fmt.Sprintf("Cannot open certificate %d.", reqId[0])
		res := Resp{msg, "Error"}
		tmpl.Execute(w, res)
		return
	}
	// Certificate doesn't exist. Let's process the CSR
	file, err = ioutil.ReadFile(csrFileName)
	if err != nil {
		msg := fmt.Sprintf("Cannot open CSR %d.", reqId[0])
		res := Resp{msg, "Error"}
		tmpl.Execute(w, res)
		return
	}
	fileInfo, _ := os.Lstat(csrFileName)
	csr, err := decodeCertRequest(string(file))
	if err != nil {
		msg := fmt.Sprintf("Cannot decode CSR %d.", reqId[0])
		res := Resp{msg, "Error"}
		tmpl.Execute(w, res)
		return
	}

	orders := getSimOrders(csr.DNSNames)

	if orders.unauthorized {
		fmt.Printf("Unauthorized will be returned.\n")
		file, _ := ioutil.ReadFile(tmplUnauthorized)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "%s\n", file)
		return
	}

	issueTime := fileInfo.ModTime().Add(orders.delay)
	if issueTime.After(time.Now()) {
		// Need to wait. Respond with 'pending'.
		fmt.Printf("Certificate will be issued issue in %s.\n", issueTime.Sub(time.Now()).String())
		res := Resp{"Taken Under Submission", "The operation completed successfully. 0x0 (WIN32: 0)"}
		tmpl.Execute(w, res)
		return
	}

	if orders.reject {
		// Certificate must be rejected
		fmt.Printf("Certificate rejected.\n")
		res := Resp{"Denied by CS simulator", "The request was denied by a certificate manager or CA administrator. 0x80094014 (-2146877420 CERTSRV_E_ADMIN_DENIED_REQUEST)"}
		tmpl.Execute(w, res)
		return
	}
	// Generate the cert and send it back

	certPem, err := c.CreateCertificatePem(csr)
	if err != nil {
		// Error
		res := Resp{"Cannot create certificate", "Error"}
		tmpl.Execute(w, res)
		return
	}
	err = ioutil.WriteFile(certFileName, []byte(certPem), 0644)
	if err != nil {
		m := "Cannot write certificate file"
		fmt.Printf("%s: %s\n", m, err.Error())
		respondError(w, m)
		return
	}

	fmt.Printf("Sending certificate:\n%s\n", certPem)
	w.Header().Add("Content-Type", "application/pkix-cert")
	fmt.Fprintf(w, "%s", certPem)
	return
}

func (c *Certserv) HandleCertnewP7b(w http.ResponseWriter, r *http.Request) {
	file, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		respondError(w, "Cannot find root CA cert.")
		return
	}
	w.Header().Add("Content-Type", "application/x-pkcs7-certificates")
	fmt.Fprintf(w, "%s", file)
}

func (c *Certserv) HandleCertcarcAsp(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles(tmplCertCaRc)
	type Resp struct {
		Renewals string
	}
	res := Resp{"0"}

	tmpl.Execute(w, res)
}

func (c *Certserv) HandleCertfnshAsp(w http.ResponseWriter, req *http.Request) {
	fmt.Printf("HandleCertfnshAsp\n")
	if req.Method != "POST" {
		fmt.Printf("Received Request: %v\n", req)
		fmt.Printf("Ignoring %s\n", req.Method)
		fmt.Fprintf(w, "")
		return
	}

	err := req.ParseForm()
	if err != nil {
		m := "Cannot parse parameters"
		fmt.Printf("%s: %s\n", m, err.Error())
		respondError(w, m)
		return
	}

	bodyCsr := req.PostForm["CertRequest"]
	if bodyCsr == nil {
		fmt.Printf("Received Request: %v\n", req)
		respondError(w, "No CertRequest found")
		return
	}

	csr, err := decodeCertRequest(bodyCsr[0])
	if err != nil {
		m := "Cannot decode CSR"
		fmt.Printf("%s: %s\n", m, err.Error())
		respondError(w, m)
		return
	}

	orders := getSimOrders(csr.DNSNames)
	// Cinek
	fmt.Printf("Orders: %v\n", orders)

	if orders.unauthorized {
		fmt.Printf("Unauthorized will be returned.\n")
		file, _ := ioutil.ReadFile(tmplUnauthorized)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "%s\n", file)
		return
	}

	if orders.delay > 0 || orders.reject {
		certId := atomic.AddUint64(&c.currentID, 1)
		csrFileName := fmt.Sprintf("ca/%d.csr", certId)
		err = ioutil.WriteFile(csrFileName, []byte(bodyCsr[0]), 0644)
		if err != nil {
			m := "Cannot write CSR file"
			fmt.Printf("%s: %s\n", m, err.Error())
			respondError(w, m)
			return
		}
		tmpl, _ := template.ParseFiles(tmplCertFnsh)
		type Resp struct {
			ReqID string
		}
		res := Resp{fmt.Sprintf("%d", certId)}
		tmpl.Execute(w, res)
		return
	}

	// No delay nor rejection, so send the certificate immediately
	certPem, err := c.CreateCertificatePem(csr)
	if err != nil {
		m := "Cannot create certificate"
		fmt.Printf("%s: %s\n", m, err.Error())
		respondError(w, m)
		return
	}
	fmt.Printf("Sending certificate:\n%s\n", certPem)
	w.Header().Add("Content-Type", "application/pkix-cert")
	fmt.Fprintf(w, "%s", certPem)
}

func (c *Certserv) CreateCertificatePem(csr *x509.CertificateRequest) ([]byte, error) {

	keyUsages := x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment
	// create client certificate template
	certTemplate := &x509.Certificate{
		Signature:          csr.Signature,
		SignatureAlgorithm: csr.SignatureAlgorithm,

		PublicKeyAlgorithm: csr.PublicKeyAlgorithm,
		PublicKey:          csr.PublicKey,

		SerialNumber:   big.NewInt(mrand.Int63()),
		Issuer:         c.caCert.Issuer,
		Subject:        csr.Subject,
		NotBefore:      time.Now(),
		NotAfter:       time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:       keyUsages,
		ExtKeyUsage:    []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:       csr.DNSNames,
		EmailAddresses: csr.EmailAddresses,
		IPAddresses:    csr.IPAddresses,
		URIs:           csr.URIs,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, certTemplate, c.caCert, csr.PublicKey, c.caKey)
	if err != nil {
		return nil, fmt.Errorf("error creating x509 certificate: %s", err.Error())
	}

	/*
	   cert, err := x509.ParseCertificate(derBytes)
	   if err != nil {
	           fmt.Sprintf(m, "error decoding DER certificate bytes: %s", err.Error())
	           fmt.Printf("%s: %s\n", m, err.Error())
	           respondError(w, m)
	           return
	   }
	*/

	pemBytes := bytes.NewBuffer([]byte{})
	err = pem.Encode(pemBytes, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return nil, fmt.Errorf("error encoding certificate PEM: %s", err.Error())
	}
	return pemBytes.Bytes(), nil
}

func (c *Certserv) CreateCertificateChainPem(csr *x509.CertificateRequest) ([]byte, error) {
	bytes, err := c.CreateCertificatePem(csr)
	if err != nil {
		return nil, err
	}
	caBytes, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		return nil, fmt.Errorf("Cannot find root CA cert. %s", err.Error())
	}

	bytes = append(bytes, caBytes...)
	return bytes, nil
}

func decodeCertRequest(data string) (*x509.CertificateRequest, error) {
	block, _ := pem.Decode([]byte(data))
	var m string
	if block == nil {
		m = "Cannot decode CSR PEM"
		e := fmt.Sprintf("%s\n%s\n", m, data)
		return nil, fmt.Errorf(e)
	}

	// parse the CSR
	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		m = "Cannot parse CSR"
		e := fmt.Sprintf("%s: %s\n", m, err.Error())
		return nil, fmt.Errorf(e)
	}
	err = csr.CheckSignature()
	if err != nil {
		m = "CSR signature error"
		e := fmt.Sprintf("%s: %s\n", m, err.Error())
		return csr, fmt.Errorf(e)
	}
	return csr, nil
}

func getSimOrders(names []string) *SimOrders {
	orders := &SimOrders{false, 0, false}
	exp := regexp.MustCompile(`^([a-z0-9\-]+)\.(([a-z0-9\-]+)\.)?sim$`)

	for _, name := range names {
		found := exp.FindStringSubmatch(name)
		if len(found) > 1 {
			switch found[1] {
			case "delay":
				var err error
				orders.delay, err = time.ParseDuration(found[3])
				if err != nil {
					orders.delay = 0
					continue
				}
			case "reject":
				orders.reject = true
			case "unauthorized":
				orders.unauthorized = true
			}
		}
	}
	return orders
}

func respondError(w http.ResponseWriter, text string) {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, "%s\n", text)
}

func (c *Certserv) initRootCert() error {
	bytes, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		return fmt.Errorf("Cannot find root CA cert. %s", err.Error())
	}
	c.caCert, err = pki.DecodeX509CertificateBytes(bytes)
	if err != nil {
		return fmt.Errorf("Cannot decode root CA cert. %s", err.Error())
	}
	//fmt.Printf("%+v", c.caCert)

	bytes, err = ioutil.ReadFile(caKeyFile)
	if err != nil {
		return fmt.Errorf("Cannot find root CA key. %s", err.Error())
	}
	c.caKey, err = pki.DecodePKCS1PrivateKeyBytes(bytes)
	if err != nil {
		return fmt.Errorf("Cannot decode root CA key. %s", err.Error())
	}

	// Check for existing CSRs and update currentID.
	files, err := ioutil.ReadDir(caDir)
	if err != nil {
		return fmt.Errorf("Cannot read CA directory. %s", err.Error())
	}
	exp := regexp.MustCompile(`^([0-9]+)\.csr$`)
	var id float64
	for _, file := range files {
		found := exp.FindStringSubmatch(file.Name())
		if len(found) > 1 {
			current, err := strconv.Atoi(found[1])
			if err != nil {
				//ignore
				continue
			}
			id = math.Max(id, float64(current))
		}
	}
	c.currentID = uint64(id)
	fmt.Printf("Startign with id = %d\n", c.currentID)
	return nil
}
