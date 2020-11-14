package common

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

const (
	AccessedMsg = "accessed"
)

func GetLogAccessEntry(req *http.Request) *log.Entry {
	return log.WithField("method", req.Method).WithField("path", req.URL)
}
