package healthcheck

import (
	"github.com/golang/glog"
	"net/http"
)

func ServeHealthCheck(w http.ResponseWriter, r *http.Request) {
	glog.V(1).Info("Healthcheck passed")
	w.WriteHeader(200)
}
