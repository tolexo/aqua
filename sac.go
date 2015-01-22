package aqua

import (
	"reflect"
)

type Sac struct {
	Data map[string]interface{}
}

func NewSac() *Sac {
	return &Sac{Data: make(map[string]interface{})}
}

func (me *Sac) Set(key string, i interface{}) *Sac {

	if i == nil {
		me.Data[key] = nil
	} else {
		switch reflect.TypeOf(i).Kind() {
		case reflect.Struct:
			if s, ok := i.(Sac); ok {
				me.Data[key] = s.Data
			} else {
				panic("TODO: need to convert struct to jsonMap?")
			}
		case reflect.Map:
			me.Data[key] = i
		case reflect.Ptr:
			item := reflect.ValueOf(i).Elem().Interface()
			me.Set(key, item)
		default:
			me.Data[key] = i
		}
	}

	return me
}
