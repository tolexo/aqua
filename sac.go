package aqua

import (
	"github.com/fatih/structs"
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
				me.Data[key] = structs.Map(i)
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

// Item being merged must be a struct or a map
func (me *Sac) Merge(i interface{}) *Sac {

	switch reflect.TypeOf(i).Kind() {
	case reflect.Struct:
		if s, ok := i.(Sac); ok {
			me.Merge(s.Data)
		} else {
			me.Merge(structs.Map(i))
		}
	case reflect.Map:
		m := i.(map[string]interface{})
		for key, val := range m {
			if _, exists := me.Data[key]; exists {
				panic("Merge field already exists:" + key)
			} else {
				me.Data[key] = val
			}
		}
	default:
		panic("Can't merge something that is not struct or map")
	}

	return me
}
