package main

import (
	"flag"
	"net/http"

	"github.com/lschyi/fileServer/handler"
	handlermiddleware "github.com/lschyi/fileServer/middleware/handler"
	responsemiddleware "github.com/lschyi/fileServer/middleware/responsewriter"

	log "github.com/sirupsen/logrus"
)

type fileServerWrapper struct {
	listHandler   http.Handler
	uploadHandler http.Handler
	enableUpload  bool
}

func NewFileServer(path string, uploadable bool) (*fileServerWrapper, error) {
	fileServer := handler.NewFileServer(path)
	listHandler := handlermiddleware.NewListHandler(fileServer)
	var uploadHandler http.Handler
	var err error
	if uploadable {
		uploadHandler, err = handlermiddleware.NewUploadHandler(fileServer, path)
		if err != nil {
			return nil, err
		}
	}
	return &fileServerWrapper{
		listHandler:   listHandler,
		uploadHandler: uploadHandler,
		enableUpload:  uploadable,
	}, nil
}

func (f *fileServerWrapper) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var writer http.ResponseWriter
	writer = responsemiddleware.NewLogMiddleware(w, req)
	if f.enableUpload {
		writer = responsemiddleware.NewUploadMiddleware(writer)
		if req.Method == http.MethodPost {
			f.uploadHandler.ServeHTTP(writer, req)
		} else {
			f.listHandler.ServeHTTP(writer, req)
		}
	} else {
		f.listHandler.ServeHTTP(writer, req)
	}
}

func main() {
	port := flag.String("p", "8000", "port to serve")
	directory := flag.String("d", ".", "directory to server")
	uploadable := flag.Bool("u", true, "enable upload function")
	flag.Parse()

	server, err := NewFileServer(*directory, *uploadable)
	if err != nil {
		log.WithError(err).Fatal("can not create file server")
	}
	http.Handle("/", server)
	log.WithField("port", *port).WithField("directory", *directory).WithField("enable upload", *uploadable).Info("Serving file server")
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.WithError(err).Fatal("encounter error")
	}
}
