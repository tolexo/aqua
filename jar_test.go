package aqua

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type jarService struct {
	RestService
	echo  GetApi
	echo2 GetApi
}

func (u *jarService) Echo(j Jar) string {
	j.LoadVars()
	return j.QueryVars["abc"]
}

func (u *jarService) Echo2(j Jar) string {
	return j.QueryVars["def"]
}

func TestJarForHttpGETMethod(t *testing.T) {

	s := NewRestServer()
	s.AddService(&jarService{})
	port := getUniquePortForTestCase()
	s.RunWith(port, false)

	Convey("Given a RestServer and a service", t, func() {
		Convey("Echo service should return Query String assigned to key: abc", func() {
			url := fmt.Sprintf("http://localhost:%d/jar/echo?abc=whatsUp", port)
			_, _, content := getUrl(url, nil)
			So(content, ShouldEqual, "whatsUp")
		})
		Convey("Echo2 service should fail since LoadVars is not invoked", func() {
			url := fmt.Sprintf("http://localhost:%d/jar/echo2?def=hello", port)
			_, _, content := getUrl(url, nil)
			So(content, ShouldEqual, "")
		})

	})
}
