package healthcheck

import (
	"github.com/golang/glog"
	"net/http"
)

func HealthCheck(r *http.Request) error {
	glog.V(1).Info("Healthcheck passed")
	return nil
}
