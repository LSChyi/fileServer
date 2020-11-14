package main

import (
	"flag"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/lschyi/fileServer/middleware"

	log "github.com/sirupsen/logrus"
)

type fileServerWrapper struct {
	handler http.Handler
	path    string
}

func NewFileServer(path string) (*fileServerWrapper, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return nil, err
	}
	return &fileServerWrapper{
		handler: http.FileServer(http.Dir(dir)),
		path:    path,
	}, nil
}

func (f *fileServerWrapper) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		if entry, err := f.handlUpload(w, req); err != nil {
			entry.WithError(err).Errorf("can not upload file")
			return
		}
	}
	var writer http.ResponseWriter
	writer = middleware.NewLogMiddleware(w, req)
	writer = middleware.NewUploadMiddleware(writer)
	f.handler.ServeHTTP(writer, req)
}

func (f *fileServerWrapper) handlUpload(w http.ResponseWriter, req *http.Request) (*log.Entry, error) {
	//entry := getAccessLogEntry(req)
	entry := log.WithField("test", "test")
	file, header, err := extractForm(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return entry, err
	}
	defer file.Close()
	entry = entry.WithField("filename", header.Filename)

	finalPath := filepath.Clean(filepath.Join(f.path, req.URL.Path))
	if _, err := filepath.Rel(f.path, finalPath); err != nil {
		http.Error(w, "can not upload to the path", http.StatusUnprocessableEntity)
		return entry, err
	}
	entry.Info("try uplading file")

	if err := saveFile(finalPath, header, file, entry); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return entry, err
	}
	return nil, nil
}

func extractForm(req *http.Request) (multipart.File, *multipart.FileHeader, error) {
	if err := req.ParseMultipartForm(32 << 20); err != nil {
		return nil, nil, err
	}
	return req.FormFile("file")
}

func saveFile(path string, header *multipart.FileHeader, file multipart.File, entry *log.Entry) error {
	dstPath := filepath.Join(path, header.Filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return err
	}
	entry.WithField("file path", dstPath).Info("file uploaded")

	return nil
}

func main() {
	port := flag.String("p", "8000", "port to serve")
	directory := flag.String("d", ".", "directory to server")
	flag.Parse()

	server, err := NewFileServer(*directory)
	if err != nil {
		log.WithError(err).Fatal("can not create file server")
	}
	http.Handle("/", server)
	log.WithField("port", *port).WithField("directory", *directory).Info("Serving file server")
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.WithError(err).Fatal("encounter error")
	}
}
