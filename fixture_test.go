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
		d := Fixture{Root: "root4", Url: "path4", Version: "version4", Pretty: "false"}
		e := Fixture{Root: "root5", Url: "path5", Version: "version5", Pretty: "true", Vendor: "vnd.app5"}
		f := Fixture{Root: "root6", Url: "path6", Version: "version6", Pretty: "true", Vendor: "vnd.app6", Prefix: "pre6"}
		g := Fixture{Root: "root8", Url: "path7", Version: "version7", Pretty: "true", Vendor: "vnd.app7", Prefix: "pre7", Modules: "mod7"}
		h := Fixture{Root: "root9", Url: "path8", Version: "version8", Pretty: "true", Vendor: "vnd.app8", Prefix: "pre8", Modules: "mod8", Cache: "cache8"}
		i := Fixture{Root: "root10", Url: "path9", Version: "version9", Pretty: "true", Vendor: "vnd.app9", Prefix: "pre9", Modules: "mod9", Cache: "cache9", Ttl: "ttl9"}

		Convey("Then resolveInOrder() should pick the first non empty value from fixtures in the given order", func() {
			z := resolveInOrder(a, b, c, d, e, f, g, h, i)
			So(z.Root, ShouldEqual, a.Root)
			So(z.Url, ShouldEqual, b.Url)
			So(z.Version, ShouldEqual, c.Version)
			So(z.Pretty, ShouldEqual, d.Pretty)
			So(z.Vendor, ShouldEqual, e.Vendor)
			So(z.Prefix, ShouldEqual, f.Prefix)
			So(z.Modules, ShouldEqual, g.Modules)
			So(z.Cache, ShouldEqual, h.Cache)
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
			So(f.Vendor, ShouldEqual, "vnd.myapp")
			So(f.Prefix, ShouldEqual, "api")
			So(f.Modules, ShouldEqual, "m")
		})
	})
}
