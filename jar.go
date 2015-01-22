package aqua

import (
	"net/http"
	"strings"
)

type Jar struct {
	Posted map[string]string
	QryStr map[string]string
}

func NewJar(r *http.Request) Jar {
	out := Jar{
		Posted: make(map[string]string),
		QryStr: make(map[string]string),
	}

	r.ParseForm()
	for k, _ := range r.PostForm {
		out.Posted[k] = strings.Join(r.PostForm[k], ",")
	}
	for k, _ := range r.Form {
		// For a cleaner separation, remove
		// post vars from QryStr
		if _, found := out.Posted[k]; !found {
			out.QryStr[k] = strings.Join(r.Form[k], ",")
		}
	}

	// TODO: handle multipart form (files)

	// USED FOR DEBUGGING
	// fmt.Println("{ --- JAR --- ")
	// fmt.Printf("post:%s\n", out.Posted)
	// fmt.Printf("qstr:%s\n", out.QryStr)
	// fmt.Println("}")

	return out
}
