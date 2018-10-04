package aqua

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/tolexo/aero/panik"
)

//statusWriter implemented http ResponseWriter
type statusWriter struct {
	http.ResponseWriter
	status int
}

//WriteHeader implementing http ResponseWriter method
func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func ModAccessLog(path string) func(http.Handler) http.Handler {

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	panik.On(err)
	l := log.New(f, "", log.LstdFlags)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapedWriter := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(wrapedWriter, r)
			l.Printf("%s %s %v %.3f", r.Method, r.RequestURI, wrapedWriter.status, time.Since(start).Seconds())
		})
	}
}

func ModSlowLog(path string, msec int) func(http.Handler) http.Handler {

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	panik.On(err)
	l := log.New(f, "", log.LstdFlags)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			dur := time.Since(start).Seconds() - float64(msec)/1000.0
			if dur > 0 {
				l.Printf("%s %s %.3f", r.Method, r.RequestURI, time.Since(start).Seconds())
			}
		})
	}
}
