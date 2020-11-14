package handler

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/lschyi/fileServer/common"
	"github.com/lschyi/fileServer/handler"

	log "github.com/sirupsen/logrus"
)

type UploadHandler struct {
	handler.FileServer
	path string
}

func NewUploadHandler(fileServer handler.FileServer, path string) (*UploadHandler, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return nil, err
	}
	return &UploadHandler{
		FileServer: fileServer,
		path:       dir,
	}, nil
}

func (u *UploadHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	entry := common.GetLogAccessEntry(req)
	handler := &logUploadHandler{
		entry: entry,
		path:  u.path,
	}
	if err := handler.Handle(w, req); err != nil {
		entry.WithError(err).Errorf("can not upload file")
		return
	}
	u.FileServer.ServeHTTP(w, req)
}

type logUploadHandler struct {
	entry *log.Entry
	path  string // the accessible top most path
}

func (l *logUploadHandler) Handle(w http.ResponseWriter, req *http.Request) error {
	file, header, err := l.extractForm(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return err
	}
	defer file.Close()
	l.entry = l.entry.WithField("filename", header.Filename)

	finalPath := filepath.Clean(filepath.Join(l.path, req.URL.Path))
	if err := l.validatePath(finalPath); err != nil {
		http.Error(w, "can not upload to the path", http.StatusUnprocessableEntity)
		return err
	}
	l.entry.Info("try uploading file")

	if err := l.saveFile(finalPath, file, header); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	return nil
}

func (l *logUploadHandler) extractForm(req *http.Request) (multipart.File, *multipart.FileHeader, error) {
	if err := req.ParseMultipartForm(32 << 20); err != nil {
		return nil, nil, err
	}
	return req.FormFile("file") // FIXME should share the form key
}

func (l *logUploadHandler) validatePath(path string) error {
	if _, err := filepath.Rel(l.path, path); err != nil {
		return err
	}
	return nil
}

func (l *logUploadHandler) saveFile(path string, file multipart.File, header *multipart.FileHeader) error {
	dstPath := filepath.Join(path, header.Filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return err
	}
	l.entry.WithField("file path", dstPath).Info("file uploaded")

	return nil
}
