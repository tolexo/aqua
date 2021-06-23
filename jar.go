package aqua

import (
	_ "fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

var separator string = ","

// Jar allows access to 1/ the reqeust, and 2/ variables (post, get and body)
// However, to load these variable values, LoadVars() method needs to be invoked
// in prior.
type Jar struct {
	// request handle
	Request *http.Request

	// variables
	PostVars  map[string]string
	QueryVars map[string]string
	Body      string
}

func NewJar(r *http.Request) Jar {
	return Jar{
		Request: r,
	}
}

func (j *Jar) LoadVars() {

	if j.PostVars != nil {
		panic("Jar.Load can be called only once")
	} else {
		j.PostVars = make(map[string]string)
		j.QueryVars = make(map[string]string)
	}

	if j.Request.Method == "POST" || j.Request.Method == "PUT" || j.Request.Method == "DELETE" {
		ctype := j.Request.Header.Get("Content-Type")
		switch {
		case ctype == "application/x-www-form-urlencoded":
			j.Request.ParseForm()
			j.loadPostVars(j.Request)
			j.loadQueryVars(j.Request, true)
		case strings.HasPrefix(ctype, "multipart/form-data;"):
			// ParseMultiPart form should ideally populate
			// r.PostForm, but instead it fills r.Form
			// https://github.com/golang/go/issues/9305
			j.Request.ParseMultipartForm(1024 * 1024)
			j.loadPostVars(j.Request)
			j.loadQueryVars(j.Request, true)
		default:
			j.Body = getBody(j.Request)
		}
	} else if j.Request.Method == "GET" {
		j.Request.ParseForm()
		j.loadQueryVars(j.Request, false)
	}
}

func getBody(r *http.Request) string {
	b, err := ioutil.ReadAll(r.Body)
	panicIf(err)
	defer r.Body.Close()
	return string(b)
}

func (j *Jar) loadPostVars(r *http.Request) {
	for k, _ := range r.PostForm {
		j.PostVars[k] = strings.Join(r.PostForm[k], separator)
	}
}

func (j *Jar) loadQueryVars(r *http.Request, skipPostVars bool) {
	for k, _ := range r.Form {
		if skipPostVars {
			// only add to query-vars if it is NOT a post var
			if _, found := j.PostVars[k]; !found {
				j.QueryVars[k] = strings.Join(r.Form[k], separator)
			}
		} else {
			j.QueryVars[k] = strings.Join(r.Form[k], separator)
		}
	}

}
