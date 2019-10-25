package certserv

import (
	"time"
)

type Cert struct {
	Crt      string
	Csr      string
	Deny     bool
	Denied   bool
	SignTime time.Time
}

func (c *Cert) TimeToSign() bool {
	if time.Now().After(c.SignTime) {
		return true
	}
	return false
}
