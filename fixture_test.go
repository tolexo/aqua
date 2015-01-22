package aqua

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestFixtureResolveInOrder(t *testing.T) {

	Convey("Given different Fixtures", t, func() {
		a := Fixture{Root: "root1"}
		b := Fixture{Root: "root2", Url: "path2"}
		c := Fixture{Root: "root3", Url: "path3", Version: "version3"}
		d := Fixture{Root: "root4", Url: "path4", Version: "version4", Pretty: "true"}
		e := Fixture{Root: "root5", Url: "path5", Version: "version5", Pretty: "false", Vnd: "vnd.app"}
		f := Fixture{Root: "root6", Url: "path6", Version: "version6", Pretty: "true", Vnd: "vnd.app6", Prefix: "pre"}
		g := Fixture{Root: "root7", Url: "path7", Version: "version7", Pretty: "true", Vnd: "vnd.app7 ", Prefix: "pre7", Modules: "mod7"}

		Convey("Then resolveInOrder() should pick the first non empty value from fixtures in the given order", func() {
			z := resolveInOrder(a, b, c, d, e, f, g)
			So(z.Root, ShouldEqual, a.Root)
			So(z.Url, ShouldEqual, b.Url)
			So(z.Version, ShouldEqual, c.Version)
			So(z.Pretty, ShouldEqual, d.Pretty)
			So(z.Vnd, ShouldEqual, e.Vnd)
			So(z.Prefix, ShouldEqual, f.Prefix)
			So(z.Modules, ShouldEqual, g.Modules)
		})
	})
}

func TestNewFixtureFromTag(t *testing.T) {

	Convey("Given a struct with fields containing values in tags", t, func() {

		type aStruct struct {
			aField string `root:"home" url:"index.json" pretty:"true" version:"1.1" vnd:"vnd.myapp" prefix:"api" modules:"m"`
		}
		a := aStruct{}

		Convey("Then a new fixture can be initialized to these values", func() {
			f := NewFixtureFromTag(&a, "aField")
			So(f.Root, ShouldEqual, "home")
			So(f.Url, ShouldEqual, "index.json")
			So(f.Pretty, ShouldEqual, "true")
			So(f.Version, ShouldEqual, "1.1")
			So(f.Vnd, ShouldEqual, "vnd.myapp")
			So(f.Prefix, ShouldEqual, "api")
			So(f.Modules, ShouldEqual, "m")
		})
	})
}
