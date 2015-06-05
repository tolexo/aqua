package aqua

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type stubService struct {
	RestService
	mock       GetApi `stub:"mocks/mock.json"`
	mockNoFile GetApi `stub:"mocks/missing.json"`
}

func TestStubFileMissing(t *testing.T) {

	s := NewRestServer()
	s.AddService(&stubService{})
	port := getUniquePortForTestCase()
	s.RunWith(port, false)

	Convey("Given a service stub", t, func() {
		Convey("And when the corresponding stub file is missing in current AND executable dir", func() {
			Convey("Then the server should return 400 series error", func() {
				url := fmt.Sprintf("http://localhost:%d/stub/mock-no-file", port)
				code, _, content := getUrl(url, nil)
				So(code, ShouldEqual, 400)
				fmt.Println(content)
			})
		})
	})

}

func TestMockStub(t *testing.T) {

	s := NewRestServer()
	s.AddService(&stubService{})
	port := getUniquePortForTestCase()
	s.RunWith(port, false)

	Convey("Given a service stub", t, func() {
		Convey("And when the corresponding stub file is found in current OR executable dir", func() {
			Convey("Then the server should return content of file", func() {
				url := fmt.Sprintf("http://localhost:%d/stub/mock", port)
				_, _, content := getUrl(url, nil)
				So(content, ShouldEqual, "MOCK DATA")
			})
		})
	})

}
