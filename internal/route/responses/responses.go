package responses

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/a-h/templ"
)

func addSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Security-Policy", "default-src 'self'")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
}

func DownloadFromReader(w http.ResponseWriter, filename string, r io.Reader) {
	addSecurityHeaders(w)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	_, _ = io.Copy(w, r)
}

func JSON(w http.ResponseWriter, body any) {
	if b, err := json.Marshal(body); err != nil {
		InternalServerError(w, err)
		return
	} else {
		JSONBytes(w, b)
	}
}

func JSONBytes(w http.ResponseWriter, body []byte) {
	addSecurityHeaders(w)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write(body)
}

func Plaintext(w http.ResponseWriter, body string) {
	addSecurityHeaders(w)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(body))
}

func RenderView(w http.ResponseWriter, ctx context.Context, component templ.Component) {
	addSecurityHeaders(w)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = component.Render(ctx, w)

}

func Created(w http.ResponseWriter, location string) {
	addSecurityHeaders(w)
	w.Header().Set("Location", location)
	w.WriteHeader(http.StatusCreated)
}

func Accepted(w http.ResponseWriter) {
	addSecurityHeaders(w)
	w.WriteHeader(http.StatusAccepted)
}

func NoContent(w http.ResponseWriter) {
	addSecurityHeaders(w)
	w.WriteHeader(http.StatusNoContent)
}

func Found(w http.ResponseWriter, location string) {
	addSecurityHeaders(w)
	w.Header().Set("Location", location)
	w.WriteHeader(http.StatusFound)
}

func SeeOther(w http.ResponseWriter, location string) {
	addSecurityHeaders(w)
	w.Header().Set("Location", location)
	w.WriteHeader(http.StatusSeeOther)
}

func BadRequest(w http.ResponseWriter, err error) {
	log.Printf("error: %v", err)
	addSecurityHeaders(w)
	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

func BadRequestWithMessage(w http.ResponseWriter, message string) {
	addSecurityHeaders(w)
	http.Error(w, message, http.StatusBadRequest)
}

func Unauthorized(w http.ResponseWriter) {
	addSecurityHeaders(w)
	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}

func Forbidden(w http.ResponseWriter, err error) {
	log.Printf("error: %v", err)
	addSecurityHeaders(w)
	http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
}

func NotFound(w http.ResponseWriter) {
	addSecurityHeaders(w)
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

func InternalServerError(w http.ResponseWriter, err error) {
	log.Printf("error: %v", err)
	addSecurityHeaders(w)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
