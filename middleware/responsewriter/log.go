package responsewriter

import (
	"fmt"
	"net/http"

	"github.com/lschyi/fileServer/common"
)

type LogMiddleware struct {
	http.ResponseWriter
	req *http.Request
}

func NewLogMiddleware(writer http.ResponseWriter, req *http.Request) *LogMiddleware {
	return &LogMiddleware{
		ResponseWriter: writer,
		req:            req,
	}
}

func (l *LogMiddleware) Write(b []byte) (int, error) {
	n, err := l.ResponseWriter.Write(b)
	if err != nil {
		l.logAccessError(err)
	}
	return n, err
}

func (l *LogMiddleware) WriteHeader(statusCode int) {
	if statusCode >= http.StatusBadRequest {
		l.logAccessError(fmt.Errorf("status code set to %d", statusCode))
	}
	l.ResponseWriter.WriteHeader(statusCode)
}

func (l *LogMiddleware) logAccessError(err error) {
	common.GetLogAccessEntry(l.req).WithError(err).Error("accessed with error")
}
