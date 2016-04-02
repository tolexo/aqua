package aqua

import (
	"github.com/tolexo/aero/auth"
	"github.com/tolexo/aero/panik"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func ModAccessLog(path string) func(http.Handler) http.Handler {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	panik.On(err)
	l := log.New(f, "", log.LstdFlags)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			l.Printf("%s %s %.3f", r.Method, r.RequestURI, time.Since(start).Seconds())
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

func ModAuth() func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			code, s := auth.CheckAuth(r)
			if code == 200 {
				r.Header.Set("ModSession", s)
				next.ServeHTTP(w, r)
			} else {
				errMessage := auth.GetAuthenticationError(code)
				w.Header().Set("Content-Length", strconv.Itoa(len(errMessage)))
				w.WriteHeader(403)
				w.Write(errMessage)
			}
		})
	}
}
