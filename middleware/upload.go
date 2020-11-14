package middleware

import (
	"net/http"
)

type UploadMiddleware struct {
	origWriter   http.ResponseWriter
	formInjected bool
}

func NewUploadMiddleware(writer http.ResponseWriter) *UploadMiddleware {
	return &UploadMiddleware{
		origWriter: writer,
	}
}

func (u *UploadMiddleware) Header() http.Header {
	return u.origWriter.Header()
}

func (u *UploadMiddleware) Write(b []byte) (int, error) {
	if !u.formInjected {
		u.injectFrom()
	}
	return u.origWriter.Write(b)
}

func (u *UploadMiddleware) WriteHeader(statusCode int) {
	u.origWriter.WriteHeader(statusCode)
}

func (u *UploadMiddleware) injectFrom() {
	u.formInjected = true
	form := `<hr />
	<form enctype="multipart/form-data" method="post">
		<input name="file" type="file" />
		<input type="submit" value="upload" />
	</form>
<hr />`
	u.origWriter.Write([]byte(form))
}
