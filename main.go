package main

import (
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

type customResponseWriter struct {
	origWriter   http.ResponseWriter
	statusCode   int
	req          *http.Request
	formInjected bool
}

func (c *customResponseWriter) Header() http.Header {
	return c.origWriter.Header()
}

func (c *customResponseWriter) injectForm() {
	c.formInjected = true
	form := `<hr />
	<form enctype="multipart/form-data" method="post">
		<input name="file" type="file" />
		<input type="submit" value="upload" />
	</form>
<hr />`
	c.origWriter.Write([]byte(form))
}

func (c *customResponseWriter) Write(b []byte) (int, error) {
	if !c.formInjected {
		c.injectForm()
	}
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
	path    string
}

func NewFileServer(path string) *fileServerWrapper {
	return &fileServerWrapper{
		handler: http.FileServer(http.Dir(path)),
		path:    path,
	}
}

func (f *fileServerWrapper) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		if err := f.handlUpload(w, req); err != nil {
			log.WithError(err).Errorf("can not upload file")
			return
		}
	}
	writer := &customResponseWriter{origWriter: w, req: req}
	f.handler.ServeHTTP(writer, req)
}

func (f *fileServerWrapper) handlUpload(w http.ResponseWriter, req *http.Request) error {
	file, header, err := extractForm(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return err
	}
	defer file.Close()

	if err := saveFile(filepath.Join(f.path, req.URL.Path), header, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

func extractForm(req *http.Request) (multipart.File, *multipart.FileHeader, error) {
	if err := req.ParseMultipartForm(32 << 20); err != nil {
		return nil, nil, err
	}
	return req.FormFile("file")
}

func saveFile(path string, header *multipart.FileHeader, file multipart.File) error {
	dst, err := os.Create(filepath.Join(path, header.Filename))
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return err
	}

	return nil
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
