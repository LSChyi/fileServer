package main

import (
	"flag"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type customResponseWriter struct {
	origWriter http.ResponseWriter
	statusCode int
	req        *http.Request
}

func (c *customResponseWriter) Header() http.Header {
	return c.origWriter.Header()
}

func (c *customResponseWriter) Write(b []byte) (int, error) {
	n, err := c.origWriter.Write(b)
	c.logAccessError(err)
	return n, err
}

func (c *customResponseWriter) WriteHeader(statusCode int) {
	c.statusCode = statusCode
	if statusCode >= http.StatusBadRequest {
		c.logAccessError(fmt.Errorf("status code set to %d", statusCode))
	} else {
		c.logAccess()
	}
	c.origWriter.WriteHeader(statusCode)
}

func (c *customResponseWriter) getLogEntry() *log.Entry {
	l := log.NewEntry(log.StandardLogger())
	if c.req != nil {
		l = log.WithField("method", c.req.Method).WithField("path", c.req.URL)
	}
	return l
}

func (c *customResponseWriter) logAccess() {
	c.getLogEntry().Info("accessed")
}

func (c *customResponseWriter) logAccessError(err error) {
	if err == nil {
		return
	}
	c.getLogEntry().WithError(err).Error("accessed with error")
}

type fileServerWrapper struct {
	handler http.Handler
}

func NewFileServer(path string) *fileServerWrapper {
	return &fileServerWrapper{
		handler: http.FileServer(http.Dir(path)),
	}
}

func (f *fileServerWrapper) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	writer := &customResponseWriter{origWriter: w, req: req}
	f.handler.ServeHTTP(writer, req)
}

func main() {
	port := flag.String("p", "8000", "port to serve")
	directory := flag.String("d", ".", "directory to server")
	flag.Parse()

	http.Handle("/", NewFileServer(*directory))
	log.WithField("port", *port).WithField("directory", *directory).Info("Serving file server")
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.WithError(err).Fatal("encounter error")
	}
}
