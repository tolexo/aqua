package aqua

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/carbocation/interpose"
	"github.com/gorilla/mux"
	"github.com/tolexo/aero/cache"
	monit "github.com/tolexo/aero/monit"
	"github.com/tolexo/aero/panik"
	"net/http"
	"reflect"
	"strings"
	"time"
)

type endPoint struct {
	caller     MethodInvoker
	info       Fixture
	httpMethod string

	isStdHttpHandler bool
	needsJarInput    bool

	muxUrl    string
	muxVars   []string
	modules   []func(http.Handler) http.Handler
	stash     cache.Cacher
	serviceId string
}

func NewEndPoint(inv MethodInvoker, f Fixture, matchUrl string, httpMethod string, mods map[string]func(http.Handler) http.Handler,
	caches map[string]cache.Cacher, serviceId string) endPoint {

	out := endPoint{
		caller:           inv,
		info:             f,
		isStdHttpHandler: false,
		needsJarInput:    false,
		muxUrl:           matchUrl,
		muxVars:          extractRouteVars(matchUrl),
		httpMethod:       httpMethod,
		modules:          make([]func(http.Handler) http.Handler, 0),
		stash:            nil,
		serviceId:        serviceId,
	}

	if f.Stub == "" {
		out.isStdHttpHandler = out.signatureMatchesDefaultHttpHandler()
		out.needsJarInput = out.needsVariableJar()

		out.validateMuxVarsMatchFuncInputs()
		out.validateFuncInputsAreOfRightType()
		out.validateFuncOutputsAreCorrect()
	}

	// Tag modules used by this endpoint
	if mods != nil && f.Modules != "" {
		names := strings.Split(f.Modules, ",")
		out.modules = make([]func(http.Handler) http.Handler, 0)
		for _, name := range names {
			name = strings.TrimSpace(name)
			fn, found := mods[name]
			if !found {
				panic(fmt.Sprintf("Module:%s not found", name))
			}
			out.modules = append(out.modules, fn)
		}
	}

	// Tag the cache
	if c, ok := caches[f.Cache]; ok {
		out.stash = c
	} else if f.Cache != "" {
		panic("Cache not found: " + f.Cache + " for " + matchUrl)
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
		if me.caller.inpParams[i] == "st:github.com/tolexo/aqua.Jar" {
			panic("Jar parameter should be the last one: " + me.caller.name)
		}
	}
	return me.caller.inpCount > 0 && me.caller.inpParams[me.caller.inpCount-1] == "st:github.com/tolexo/aqua.Jar"
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
			case "st:github.com/tolexo/aqua.Jar":
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
	accepts["st:github.com/tolexo/aqua.Sac"] = true
	accepts["*st:github.com/tolexo/aqua.Sac"] = true

	if !me.isStdHttpHandler {
		switch me.caller.outCount {
		case 1:
			_, found := accepts[me.caller.outParams[0]]
			if !found && !strings.HasPrefix(me.caller.outParams[0], "st:") {
				fmt.Println(me.caller.outParams[0])
				panic("Incorrect return type found in: " + me.caller.name)
			}
		case 2:
			if me.caller.outParams[0] != "int" {
				panic("When a func returns two params, the first must be an int (http status code) : " + me.caller.name)
			}
			_, found := accepts[me.caller.outParams[1]]
			if !found && !strings.HasPrefix(me.caller.outParams[1], "st:") {
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
		//fmt.Println("using module:", me.modules[i], reflect.TypeOf(me.modules[i]))
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
		header1 := fmt.Sprintf("application/%s-v%s+json", me.info.Vendor, me.info.Version)
		mux.Handle(urlWithoutVersion, m).Methods(me.httpMethod).Headers("Accept", header1)

		// content type (style2)
		header2 := fmt.Sprintf("application/%s+json;version=%s", me.info.Vendor, me.info.Version)
		mux.Handle(urlWithoutVersion, m).Methods(me.httpMethod).Headers("Accept", header2)
	}
}

func handleIncoming(e *endPoint) func(http.ResponseWriter, *http.Request) {

	// return stub
	if e.info.Stub != "" {
		return func(w http.ResponseWriter, r *http.Request) {
			d, err := getContent(e.info.Stub)
			if err == nil {
				fmt.Fprintf(w, "%s", d)
			} else {
				w.WriteHeader(400)
				fmt.Fprintf(w, "{ message: \"%s\"}", "Stub path not found")
			}
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {

		cacheHit := false

		// TODO: create less local variables
		// TODO: move vars to closure level

		var out []reflect.Value

		//TODO: capture this using instrumentation handler
		defer func(reqStartTime time.Time) {
			go func() {
				if e.serviceId != "" {
					respTime := time.Since(reqStartTime).Seconds() * 1000
					var responseCode int64 = 200
					if out != nil && len(out) == 2 && e.caller.outParams[0] == "int" {
						responseCode = out[0].Int()
					}
					monitorParams := monit.MonitorParams{
						ServiceId:    e.serviceId,
						RespTime:     respTime,
						ResponseCode: responseCode,
						CacheHit:     cacheHit,
					}
					monit.MonitorMe(monitorParams)
				}
			}()
		}(time.Now())

		var useCache bool = false
		var ttl time.Duration = 0 * time.Second
		var val []byte
		var err error

		if e.info.Ttl != "" {
			ttl, err = time.ParseDuration(e.info.Ttl)
			panik.On(err)
		}
		useCache = r.Method == "GET" && ttl > 0 && e.stash != nil

		muxVals := mux.Vars(r)
		params := make([]string, len(e.muxVars))
		for i, v := range e.muxVars {
			params[i] = muxVals[v]
		}

		if e.isStdHttpHandler {
			//TODO: caching of standard handler
			e.caller.Do([]reflect.Value{reflect.ValueOf(w), reflect.ValueOf(r)})
		} else {
			ref := convertToType(params, e.caller.inpParams)
			if e.needsJarInput {
				ref = append(ref, reflect.ValueOf(NewJar(r)))
			}

			if useCache {
				val, err = e.stash.Get(r.RequestURI)
				if err == nil {
					cacheHit = true
					// fmt.Print(".")
					out = decomposeCachedValues(val, e.caller.outParams)
				} else {
					out = e.caller.Do(ref)
					if len(out) == 2 && e.caller.outParams[0] == "int" {
						code := out[0].Int()
						if code < 200 || code > 299 {
							useCache = false
						}
					}
					if useCache {
						bytes := prepareForCaching(out, e.caller.outParams)
						e.stash.Set(r.RequestURI, bytes, ttl)
						// fmt.Print(":", len(bytes), r.RequestURI)
					}

				}
			} else {
				out = e.caller.Do(ref)
				// fmt.Print("!")
			}
			writeOutput(w, e.caller.outParams, out, e.info.Pretty)
		}
	}
}

func prepareForCaching(r []reflect.Value, outputParams []string) []byte {

	var err error
	buf := new(bytes.Buffer)
	encd := json.NewEncoder(buf)

	for i, _ := range r {
		switch outputParams[i] {
		case "int":
			err = encd.Encode(r[i].Int())
			panik.On(err)
		case "map":
			err = encd.Encode(r[i].Interface().(map[string]interface{}))
			panik.On(err)
		case "string":
			err = encd.Encode(r[i].String())
			panik.On(err)
		case "*st:github.com/tolexo/aqua.Sac":
			err = encd.Encode(r[i].Elem().Interface().(Sac).Data)
		default:
			panic("Unknown type of output to be sent to endpoint cache: " + outputParams[i])
		}
	}

	return buf.Bytes()
}

func decomposeCachedValues(data []byte, outputParams []string) []reflect.Value {

	var err error
	buf := bytes.NewBuffer(data)
	decd := json.NewDecoder(buf)
	out := make([]reflect.Value, len(outputParams))

	for i, o := range outputParams {
		switch o {
		case "int":
			var j int
			err = decd.Decode(&j)
			panik.On(err)
			out[i] = reflect.ValueOf(j)
		case "map":
			var m map[string]interface{}
			err = decd.Decode(&m)
			panik.On(err)
			out[i] = reflect.ValueOf(m)
		case "string":
			var s string
			err = decd.Decode(&s)
			panik.On(err)
			out[i] = reflect.ValueOf(s)
		case "*st:github.com/tolexo/aqua.Sac":
			var m map[string]interface{}
			err = decd.Decode(&m)
			panik.On(err)
			s := NewSac()
			s.Data = m
			out[i] = reflect.ValueOf(s)
		default:
			panic("Unknown type of output to be decoded from endpoint cache:" + o)
		}
	}

	return out

}
