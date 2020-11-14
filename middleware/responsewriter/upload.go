package responsewriter

import (
	"net/http"
	"os"
	"path/filepath"
)

type UploadMiddleware struct {
	http.ResponseWriter
	req                *http.Request
	path               string
	isInjectionHandled bool
}

func NewUploadMiddleware(writer http.ResponseWriter, req *http.Request, path string) *UploadMiddleware {
	return &UploadMiddleware{
		ResponseWriter: writer,
		req:            req,
		path:           path,
	}
}

func (u *UploadMiddleware) Write(b []byte) (int, error) {
	if !u.isInjectionHandled {
		u.isInjectionHandled = true
		if u.shouldInject() {
			u.injectFrom()
		}
	}
	return u.ResponseWriter.Write(b)
}

func (u *UploadMiddleware) WriteHeader(statusCode int) {
	u.ResponseWriter.WriteHeader(statusCode)
}

func (u *UploadMiddleware) injectFrom() {
	form := `<hr />
	<form enctype="multipart/form-data" method="post">
		<input name="file" type="file" />
		<input type="submit" value="upload" />
	</form>
<hr />`
	u.ResponseWriter.Write([]byte(form))
}

func (u *UploadMiddleware) shouldInject() bool {
	dstPath := filepath.Join(u.path, u.req.URL.Path)
	info, err := os.Stat(dstPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}
