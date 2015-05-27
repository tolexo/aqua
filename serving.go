package aqua

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

func writeOutput(w http.ResponseWriter, outType []string, outVals []reflect.Value, pretty string) {

	if len(outType) == 1 {
		if outType[0] == "int" {
			w.WriteHeader(int(outVals[0].Int()))
		} else {
			writeItem(w, outType[0], outVals[0], pretty)
		}
	} else if len(outType) == 2 {
		w.WriteHeader(int(outVals[0].Int()))
		writeItem(w, outType[1], outVals[1], pretty)
	}
}

func writeItem(w http.ResponseWriter, oType string, oVal reflect.Value, pretty string) {

	if strings.HasPrefix(oType, "*st:") {
		o := oVal.Elem()
		writeItem(w, getSymbolFromType(o.Type()), o, pretty)
		return
	}

	switch {
	case oType == "string":
		v := oVal.String()
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Length", strconv.Itoa(len(v)))
		fmt.Fprintf(w, "%s", v)
	case oType == "st:github.com/thejackrabbit/aqua.Sac":
		s := oVal.Interface().(Sac)
		writeItem(w, getSymbolFromType(reflect.TypeOf(s.Data)), reflect.ValueOf(s.Data), pretty)
	case oType == "map", strings.HasPrefix(oType, "st:"):
		var j []byte
		var e error
		if pretty == "true" || pretty == "1" {
			j, e = json.MarshalIndent(oVal.Interface(), "", "\t")
			panicIf(e)
		} else {
			j, e = json.Marshal(oVal.Interface())
			panicIf(e)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", strconv.Itoa(len(j)))
		w.Write(j)
	default:
		fmt.Printf("Don't know how to return a: %s?\n", oType)
	}
}
