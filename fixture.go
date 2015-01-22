package aqua

import (
	"reflect"
)

type Fixture struct {
	Prefix  string
	Root    string
	Url     string
	Version string
	Pretty  string
	Vnd     string
	Modules string
}

func NewFixtureFromTag(i interface{}, fieldName string) Fixture {
	out := Fixture{}
	field, _ := reflect.TypeOf(i).Elem().FieldByName(fieldName)
	tag := field.Tag

	if tag.Get("prefix") != "" {
		out.Prefix = tag.Get("prefix")
	}
	if tag.Get("root") != "" {
		out.Root = tag.Get("root")
	}
	if tag.Get("url") != "" {
		out.Url = tag.Get("url")
	}
	if tag.Get("version") != "" {
		out.Version = tag.Get("version")
	}
	if tag.Get("pretty") != "" {
		out.Pretty = tag.Get("pretty")
	}
	if tag.Get("vnd") != "" {
		out.Vnd = tag.Get("vnd")
	}
	if tag.Get("modules") != "" {
		out.Modules = tag.Get("modules")
	}

	return out
}

func resolveInOrder(e ...Fixture) Fixture {
	out := Fixture{}
	empty := ""

	for _, ep := range e {
		if out.Prefix == empty && ep.Prefix != empty {
			out.Prefix = ep.Prefix
		}
		if out.Root == empty && ep.Root != empty {
			out.Root = ep.Root
		}
		if out.Url == empty && ep.Url != empty {
			out.Url = ep.Url
		}
		if out.Version == empty && ep.Version != empty {
			out.Version = ep.Version
		}
		if out.Pretty == empty && ep.Pretty != empty {
			out.Pretty = ep.Pretty
		}
		if out.Vnd == empty && ep.Vnd != empty {
			out.Vnd = ep.Vnd
		}
		if out.Modules == empty && ep.Modules != empty {
			out.Modules = ep.Modules
		}
	}
	return out
}
