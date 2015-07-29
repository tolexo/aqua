package aqua

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestCoreFunctions(t *testing.T) {

	s := NewRestServer()
	port := getUniquePortForTestCase()
	s.Port = port
	s.RunAsync()

	Convey("When you start a RestServer", t, func() {

		Convey("Then /aqua/ping should response with pong", func() {
			url := fmt.Sprintf("http://localhost:%d/aqua/ping", port)
			code, ctype, content := getUrl(url, nil)
			So(code, ShouldEqual, 200)
			So(ctype, ShouldEqual, "text/plain")
			So(content, ShouldEqual, "pong")
		})

		Convey("Then /aqua/status should response with a json object", func() {
			url := fmt.Sprintf("http://localhost:%d/aqua/status", port)
			code, ctype, content := getUrl(url, nil)
			So(code, ShouldEqual, 200)
			So(ctype, ShouldEqual, "application/json")
			So(content, ShouldContainSubstring, "aqua-version")
			So(content, ShouldContainSubstring, "server-time")
		})

		Convey("Then /aqua/time should respond with a timestamp", func() {
			url := fmt.Sprintf("http://localhost:%d/aqua/time", port)
			code, ctype, content := getUrl(url, nil)
			So(code, ShouldEqual, 200)
			So(ctype, ShouldStartWith, "text/plain")
			_, err := time.Parse("2006-01-02 15:04:05 MST", content)
			So(err, ShouldBeNil)
		})
	})

}
