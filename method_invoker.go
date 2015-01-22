package aqua

import (
	"fmt"
	"reflect"
	"strings"
)

type MethodInvoker struct {
	o      interface{}
	name   string
	exists bool

	outCount  int
	outParams []string

	inpCount  int
	inpParams []string
}

func NewMethodInvoker(i interface{}, method string) MethodInvoker {
	out := &MethodInvoker{
		o:      i,
		name:   method,
		exists: false,
	}

	symb := getSymbolFromObject(i)
	if !strings.HasPrefix(symb, "*st:") {
		panic("MethodInvoker expects address of struct")
	}

	var m reflect.Method
	m, out.exists = reflect.TypeOf(out.o).MethodByName(out.name)
	if out.exists {
		out.decipherOutputs(m.Type)
		out.decipherInputs(m.Type)
	}

	return *out
}

func (me *MethodInvoker) decipherOutputs(mt reflect.Type) {

	me.outCount = mt.NumOut()
	me.outParams = make([]string, mt.NumOut())

	for i := 0; i < mt.NumOut(); i++ {
		pt := mt.Out(i)
		me.outParams[i] = getSymbolFromType(pt)
	}
}

func (me *MethodInvoker) decipherInputs(mt reflect.Type) {

	me.inpCount = mt.NumIn() - 1 // skip the first param (me)
	me.inpParams = make([]string, mt.NumIn()-1)

	for i := 1; i < mt.NumIn(); i++ {
		pt := mt.In(i)
		me.inpParams[i-1] = getSymbolFromType(pt)
	}
}

func (me *MethodInvoker) Do(v []reflect.Value) []reflect.Value {
	return reflect.ValueOf(me.o).MethodByName(me.name).Call(v)
}

func (me *MethodInvoker) Pr() {
	fmt.Printf("%s.%s has %d inputs and %d outParamsputs\n", me.o, me.name, me.inpCount, me.outCount)
	for i, s := range me.outParams {
		fmt.Printf(" outParams -> %s && %s\n", s, me.outParams[i])
	}
	for i, s := range me.inpParams {
		fmt.Printf(" inpParams -> %s && %s\n", s, me.inpParams[i])
	}
}
