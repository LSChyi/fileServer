package handler

import (
	"net/http"

	"github.com/lschyi/fileServer/common"
	"github.com/lschyi/fileServer/handler"
)

type ListHandler struct {
	handler.FileServer
}

func NewListHandler(fileServer handler.FileServer) *ListHandler {
	return &ListHandler{
		FileServer: fileServer,
	}
}

func (l *ListHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	common.GetLogAccessEntry(req).Info(common.AccessedMsg)
	l.FileServer.ServeHTTP(w, req)
}
