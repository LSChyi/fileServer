package middleware

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type LogMiddleware struct {
	origWriter http.ResponseWriter
	req        *http.Request
}

func NewLogMiddleware(writer http.ResponseWriter, req *http.Request) *LogMiddleware {
	return &LogMiddleware{
		origWriter: writer,
		req:        req,
	}
}

func (l *LogMiddleware) Header() http.Header {
	return l.origWriter.Header()
}

func (l *LogMiddleware) Write(b []byte) (int, error) {
	n, err := l.origWriter.Write(b)
	l.logAccessError(err)
	return n, err
}

func (l *LogMiddleware) WriteHeader(statusCode int) {
	if statusCode >= http.StatusBadRequest {
		l.logAccessError(fmt.Errorf("status code set to %d", statusCode))
	} else {
		l.logAccess()
	}
	l.origWriter.WriteHeader(statusCode)
}

func (l *LogMiddleware) logAccess() {
	l.getLogEntry().Info("accessed")
}

func (l *LogMiddleware) logAccessError(err error) {
	if err == nil {
		return
	}
	l.getLogEntry().WithError(err).Error("accessed with error")
}

func (l *LogMiddleware) getLogEntry() *log.Entry {
	return getAccessLogEntry(l.req)
}

func getAccessLogEntry(req *http.Request) *log.Entry {
	entry := log.NewEntry(log.StandardLogger())
	if req != nil {
		entry = log.WithField("method", req.Method).WithField("path", req.URL)
	}
	return entry
}
