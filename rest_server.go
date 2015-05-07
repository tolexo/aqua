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
}

var release string = "0.0.1"
var defaultPort int = 8090

type RestServer struct {
	Fixture
	http.Server
	mux  *mux.Router
	apis map[string]endPoint
	mods map[string]func(http.Handler) http.Handler
	cach map[string]cache.Cacher
}

func NewRestServer() RestServer {
	r := RestServer{
		Fixture: defaults,
		Server:  http.Server{},
		mux:     mux.NewRouter(),
		apis:    make(map[string]endPoint),
		mods:    make(map[string]func(http.Handler) http.Handler),
		cach:    make(map[string]cache.Cacher),
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
	// TODO: check if the same key alread exists
	// TODO: AddCache must be called before AddService
	me.cach[name] = c
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
		if inv.exists {
			ep := NewEndPoint(inv, fix, matchUrl, method, me.mods, me.cach)
			ep.setupMuxHandlers(me.mux)
			me.apis[serviceId] = ep
			fmt.Printf("%s\n", serviceId)
		}
	}
}

func (me *RestServer) Run() {
	me.RunWith(0, true)
}

func (me *RestServer) RunWith(port int, sync bool) {
	if sync {
		startup(me, port)
	} else {
		go startup(me, port)

		// TODO: don't sleep, check for the server to come up, and panic if
		// it doesn't even after 5 sec
		time.Sleep(time.Millisecond * 50)
	}
}

func startup(r *RestServer, port int) {
	if port > 0 {
		r.Addr = fmt.Sprintf(":%d", port)
	} else if r.Server.Addr == "" {
		r.Addr = fmt.Sprintf(":%d", defaultPort)
	}
	r.Server.Handler = r.mux
	fmt.Println(r.ListenAndServe())
}
