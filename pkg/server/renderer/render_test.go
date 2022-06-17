package renderer_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/athosone/golib/pkg/server/renderer"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type testStruct struct {
	Name string `json:"nameJson" xml:"NameXml" yaml:"nameYaml"`
}

var _ = Describe("Render", func() {
	const (
		v1beta1Json = "application/vnd.athosone.test+json; version=v1beta1"
		v1beta1Yaml = "application/vnd.athosone.test+yaml; version=v1beta1"
		v1beta1Xml  = "application/vnd.athosone.test+xml; v=v1beta1"
		v1beta2     = "application/vnd.athosone.test+xml; v=v1beta2"
		v2          = "application/vnd.athosone.test+yaml; v=v2"
	)
	var (
		responseRecorder *httptest.ResponseRecorder
		request          *http.Request
		test             testStruct
	)
	BeforeEach(func() {
		responseRecorder = httptest.NewRecorder()
		request, _ = http.NewRequest("GET", "https://example.com/", nil)
		test = testStruct{
			Name: "test",
		}
	})
	JustBeforeEach(func() {
		Expect(renderer.OK(responseRecorder, request, test)).To(Succeed())
	})
	When("Request has yaml accept header", func() {
		BeforeEach(func() {
			request.Header.Set("Accept", "application/yaml")
		})
		It("should set content type to yaml", func() {
			Expect(responseRecorder.Header().Get("Content-Type")).To(Equal("application/yaml"))
		})
		It("should return yaml", func() {
			Expect(responseRecorder.Body.String()).To(Equal("nameYaml: test\n"))
		})
	})
	When("Request has json accept header", func() {
		BeforeEach(func() {
			request.Header.Set("Accept", "application/json")
		})
		It("should set content type to json", func() {
			Expect(responseRecorder.Header().Get("Content-Type")).To(Equal("application/json"))
		})
		It("should return json", func() {
			Expect(responseRecorder.Body.String()).To(Equal("{\"nameJson\":\"test\"}\n"))
		})
	})
	When("Request has xml accept header", func() {
		BeforeEach(func() {
			request.Header.Set("Accept", "application/xml")
		})
		It("should set content type to xml", func() {
			Expect(responseRecorder.Header().Get("Content-Type")).To(Equal("application/xml"))
		})
		It("should return xml", func() {
			Expect(responseRecorder.Body.String()).To(Equal("<testStruct><NameXml>test</NameXml></testStruct>"))
		})
	})
	When("Request has v1beta1 accept header", func() {
		When("Request has v1beta1 json accept header", func() {
			BeforeEach(func() {
				request.Header.Set("Accept", v1beta1Json)
			})
			It("should set content type to json", func() {
				Expect(responseRecorder.Header().Get("Content-Type")).To(Equal(v1beta1Json))
			})
			It("should return json", func() {
				Expect(responseRecorder.Body.String()).To(Equal("{\"nameJson\":\"test\"}\n"))
			})
		})
		When("Request has v1beta1 yaml accept header", func() {
			BeforeEach(func() {
				request.Header.Set("Accept", v1beta1Yaml)
			})
			It("should set content type to yaml", func() {
				Expect(responseRecorder.Header().Get("Content-Type")).To(Equal(v1beta1Yaml))
			})
			It("should return yaml", func() {
				Expect(responseRecorder.Body.String()).To(Equal("nameYaml: test\n"))
			})
		})
		When("Request has v1beta1 xml accept header", func() {
			BeforeEach(func() {
				request.Header.Set("Accept", v1beta1Xml)
			})
			It("should set content type to xml", func() {
				Expect(responseRecorder.Header().Get("Content-Type")).To(Equal(v1beta1Xml))
			})
			It("should return xml", func() {
				Expect(responseRecorder.Body.String()).To(Equal("<testStruct><NameXml>test</NameXml></testStruct>"))
			})
		})
	})
	When("Request has v1beta1 and v1beta2 accept header with different quality", func() {
		BeforeEach(func() {
			request.Header.Set("Accept", v1beta1Json+";q=0.8"+", "+v1beta2)
		})
		It("should set content type to the most precise one (v1beta2)", func() {
			Expect(responseRecorder.Header().Get("Content-Type")).To(Equal(v1beta2))
		})
	})
	When("Request has v2 and v1beta2 accept header with same quality", func() {
		BeforeEach(func() {
			request.Header.Set("Accept", v2+";q=1.0"+", "+v1beta2)
		})
		It("should set content type to the most precise one (v2)", func() {
			Expect(responseRecorder.Header().Get("Content-Type")).To(Equal(v2))
		})
	})
	When("Request has * accept header", func() {
		BeforeEach(func() {
			request.Header.Set("Accept", "*/*")
		})
		It("should set content type to json", func() {
			Expect(responseRecorder.Header().Get("Content-Type")).To(Equal("application/json"))
		})
		It("should return json", func() {
			Expect(responseRecorder.Body.String()).To(Equal("{\"nameJson\":\"test\"}\n"))
		})
	})
})
