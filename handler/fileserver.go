package handler

import (
	"net/http"
)

type FileServer interface {
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}

type fileServer struct {
	http.Handler
}

func NewFileServer(path string) *fileServer {
	return &fileServer{
		Handler: http.FileServer(http.Dir(path)),
	}
}
