package aqua

import (
	"fmt"
	"github.com/carbocation/interpose"
	"github.com/gorilla/mux"
	"net/http"
	"reflect"
	"strings"
)

type endPoint struct {
	caller     MethodInvoker
	info       Fixture
	httpMethod string

	isStdHttpHandler bool
	needsJarInput    bool

	muxUrl  string
	muxVars []string
	modules []func(http.Handler) http.Handler
}

func NewEndPoint(inv MethodInvoker, f Fixture, matchUrl string, httpMethod string, mods map[string]func(http.Handler) http.Handler) endPoint {
	out := endPoint{
		caller:           inv,
		info:             f,
		isStdHttpHandler: false,
		needsJarInput:    false,
		muxUrl:           matchUrl,
		muxVars:          extractRouteVars(matchUrl),
		httpMethod:       httpMethod,
		modules:          make([]func(http.Handler) http.Handler, 0),
	}

	out.isStdHttpHandler = out.signatureMatchesDefaultHttpHandler()
	out.needsJarInput = out.needsVariableJar()

	out.validateMuxVarsMatchFuncInputs()
	out.validateFuncInputsAreOfRightType()
	out.validateFuncOutputsAreCorrect()

	if mods != nil && f.Modules != "" {
		names := strings.Split(f.Modules, ",")
		out.modules = make([]func(http.Handler) http.Handler, len(names))
		for _, name := range names {
			name = strings.TrimSpace(name)
			fn, found := mods[name]
			if !found {
				panic(fmt.Sprintf("Module:%s not found", name))
			}
			out.modules = append(out.modules, fn)
		}
	}

	return out
}

func (me *endPoint) signatureMatchesDefaultHttpHandler() bool {
	return me.caller.outCount == 0 &&
		me.caller.inpCount == 2 &&
		me.caller.inpParams[0] == "i:net/http.ResponseWriter" &&
		me.caller.inpParams[1] == "*st:net/http.Request"
}

func (me *endPoint) needsVariableJar() bool {
	// needs jar input as the last parameter
	for i := 0; i < len(me.caller.inpParams)-1; i++ {
		if me.caller.inpParams[i] == "st:aqua.Jar" {
			panic("Jar parameter should be the last one: " + me.caller.name)
		}
	}
	return me.caller.inpCount > 0 && me.caller.inpParams[me.caller.inpCount-1] == "st:aqua.Jar"
}

func (me *endPoint) validateMuxVarsMatchFuncInputs() {
	// for non-standard http handlers, the mux vars count should match
	// the count of inputs to the user's method
	if !me.isStdHttpHandler {
		inputs := me.caller.inpCount
		if me.needsJarInput {
			inputs += -1
		}
		if len(me.muxVars) != inputs {
			panic(fmt.Sprintf("%s has %d inputs, but the func (%s) has %d",
				me.muxUrl, len(me.muxVars), me.caller.name, inputs))
		}
	}
}

func (me *endPoint) validateFuncInputsAreOfRightType() {
	if !me.isStdHttpHandler {
		for _, s := range me.caller.inpParams {
			switch s {
			case "st:aqua.Jar":
			case "int":
			case "string":
			default:
				panic("Func input params should be 'int' or 'string'. Observed: " + s + " in: " + me.caller.name)
			}
		}
	}
}

func (me *endPoint) validateFuncOutputsAreCorrect() {

	var accepts = make(map[string]bool)
	accepts["string"] = true
	accepts["map"] = true
	accepts["st:aqua.Sac"] = true
	accepts["*st:aqua.Sac"] = true

	if !me.isStdHttpHandler {
		switch me.caller.outCount {
		case 1:
			if _, found := accepts[me.caller.outParams[0]]; !found {
				panic("Incorrect return type found in: " + me.caller.name)
			}
		case 2:
			if me.caller.outParams[0] != "int" {
				panic("When a func returns two params, the first must be an int (http status code) : " + me.caller.name)
			}
			if _, found := accepts[me.caller.outParams[1]]; !found {
				panic("Incorrect return type for second return param found in: " + me.caller.name)
			}
		default:
			panic("Incorrect number of returns for Func: " + me.caller.name)
		}
	}
}

// func middleman(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Println("In the middle >>>>")
// 		next.ServeHTTP(w, r)
// 		fmt.Println("And leaving middle <<<<")
// 	})
// }

func (me *endPoint) setupMuxHandlers(mux *mux.Router) {

	fn := handleIncoming(me)

	m := interpose.New()
	for i, _ := range me.modules {
		m.Use(me.modules[i])
	}
	m.UseHandler(http.HandlerFunc(fn))

	if me.info.Version == "*" {
		mux.Handle(me.muxUrl, m).Methods(me.httpMethod)
	} else {
		urlWithVersion := cleanUrl(me.info.Prefix, "v"+me.info.Version, me.muxUrl)
		urlWithoutVersion := cleanUrl(me.info.Prefix, me.muxUrl)

		// versioned url
		mux.Handle(urlWithVersion, m).Methods(me.httpMethod)

		// content type (style1)
		header1 := fmt.Sprintf("application/%s-v%s+json", me.info.Vnd, me.info.Version)
		mux.Handle(urlWithoutVersion, m).Methods(me.httpMethod).Headers("Accept", header1)

		// content type (style2)
		header2 := fmt.Sprintf("application/%s+json;version=%s", me.info.Vnd, me.info.Version)
		mux.Handle(urlWithoutVersion, m).Methods(me.httpMethod).Headers("Accept", header2)
	}
}

func handleIncoming(e *endPoint) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// todo: create less local variables

		var out []reflect.Value

		muxVals := mux.Vars(r)
		params := make([]string, len(e.muxVars))
		for i, v := range e.muxVars {
			params[i] = muxVals[v]
		}

		if e.isStdHttpHandler {
			e.caller.Do([]reflect.Value{reflect.ValueOf(w), reflect.ValueOf(r)})
		} else {
			ref := convertToType(params, e.caller.inpParams)
			if e.needsJarInput {
				ref = append(ref, reflect.ValueOf(NewJar(r)))
			}
			out = e.caller.Do(ref)
			writeOutput(w, e.caller.outParams, out, e.info.Pretty)
		}
	}
}
