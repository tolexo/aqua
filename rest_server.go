package aqua

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/thejackrabbit/aero/cache"
	"net/http"
	"reflect"
	"strings"
	"time"
)

var defaults Fixture = Fixture{
	Root:    "",
	Url:     "",
	Version: "*",
	Pretty:  "false",
	Vendor:  "vnd.api",
	Modules: "",
	Cache:   "",
	Ttl:     "",
}

var release string = "0.0.1"
var defaultPort int = 8090

type RestServer struct {
	Fixture
	http.Server
	Port   int
	mux    *mux.Router
	apis   map[string]endPoint
	mods   map[string]func(http.Handler) http.Handler
	stores map[string]cache.Cacher
}

func NewRestServer() RestServer {
	r := RestServer{
		Fixture: defaults,
		Server:  http.Server{},
		Port:    defaultPort,
		mux:     mux.NewRouter(),
		apis:    make(map[string]endPoint),
		mods:    make(map[string]func(http.Handler) http.Handler),
		stores:  make(map[string]cache.Cacher),
	}
	r.AddService(&CoreService{})
	return r
}

var printed bool = false

func (me *RestServer) AddModule(name string, f func(http.Handler) http.Handler) {
	// TODO: check if the same key alread exists
	// TODO: AddModule must be called before AddService
	me.mods[name] = f
}

func (me *RestServer) AddCache(name string, c cache.Cacher) {
	// TODO: check if the same key already exists
	// TODO: AddCache must be called before AddService
	me.stores[name] = c
}

func (me *RestServer) AddService(svc interface{}) {

	svcType := reflect.TypeOf(svc)
	code := getSymbolFromType(svcType)

	// validation: must be pointer
	if !strings.HasPrefix(code, "*st:") {
		panic("RestServer.AddService() expects address of your Service object")
	}

	// validation: RestService field must be present and be anonymous
	rs, ok := svcType.Elem().FieldByName("RestService")
	if !ok || !rs.Anonymous || !rs.Type.ConvertibleTo(reflect.TypeOf(RestService{})) {
		panic("RestServer.AddService() expects object that contains anonymous RestService field")
	}

	objType := svcType.Elem()
	obj := reflect.ValueOf(svc).Elem()

	fixSvcTag := NewFixtureFromTag(svc, "RestService")
	fixSvc := obj.FieldByName("RestService").FieldByName("Fixture").Interface().(Fixture)

	var fix Fixture
	var method string

	if !printed {
		fmt.Println("Loading...")
		printed = true
	}

	for i := 0; i < objType.NumField(); i++ {
		field := objType.FieldByIndex([]int{i})

		method = getHttpMethod(field)
		if method == "" {
			continue
		}
		fixFldTag := NewFixtureFromTag(svc, field.Name)
		fix = resolveInOrder(fixFldTag, fixSvc, fixSvcTag, me.Fixture)

		if fix.Root == "" {
			tmp := objType.Name()
			if strings.HasSuffix(tmp, "Service") {
				tmp = tmp[0 : len(tmp)-len("Service")]
			}
			fix.Root = toUrlCase(tmp)
		}
		if fix.Url == "" {
			fix.Url = toUrlCase(field.Name)
		}

		matchUrl := cleanUrl(fix.Root, fix.Url)
		serviceId := getServiceId(method, fix.Prefix, fix.Version, matchUrl)

		if _, found := me.apis[serviceId]; found {
			panic("Cannot load service again: " + serviceId)
		}

		inv := NewMethodInvoker(svc, upFirstChar(field.Name))
		if inv.exists || fix.Stub != "" {
			ep := NewEndPoint(inv, fix, matchUrl, method, me.mods, me.stores)
			ep.setupMuxHandlers(me.mux)
			me.apis[serviceId] = ep
			fmt.Printf("%s\n", serviceId)
			//fmt.Println(fix)
		}
	}
}

func (me *RestServer) Run() {
	startup(me, me.Port)
}

func (me *RestServer) RunAsync() {
	go startup(me, me.Port)

	// TODO: don't sleep, check for the server to come up, and panic if
	// it doesn't even after 5 sec
	time.Sleep(time.Millisecond * 50)
}

func startup(r *RestServer, port int) {
	if port > 0 {
		r.Addr = fmt.Sprintf(":%d", port)
	} else if r.Server.Addr == "" {
		r.Addr = fmt.Sprintf(":%d", port)
	}
	r.Server.Handler = r.mux
	fmt.Println(r.ListenAndServe())
}
