package routing_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/athosone/golib/pkg/server/routing"
)

var _ = Describe("Routing", func() {
	var (
		gRouter           *routing.GRouter
		req               *http.Request
		responseRecorder  *httptest.ResponseRecorder
		v1Type            = "application/vnd.athosone.innersource+json; v=1"
		v2Type            = "application/vnd.athosone.innersource+json; v=2"
		wildcardType      = "application/vnd.athosone.innersource+*"
		wildcardOnlyType  = "*"
		wildcardSlashType = "*/*"
		invalidType       = "application/vnd.athosone.innersource++json"
		unsupportedType   = "application/vnd.company.domain+xml"
	)
	BeforeEach(func() {
		gRouter = routing.NewRouter()
	})
	When("a route is created", func() {
		BeforeEach(func() {
			responseRecorder = httptest.NewRecorder()
		})
		When("providing the proper headers", func() {
			BeforeEach(func() {
				gRouter.Get(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}).Produce(v1Type, v2Type)

				gRouter.Post(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusCreated)
				}).Consume(v1Type)

				gRouter.Post(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusCreated)
				}).Consume(v2Type)

				gRouter.Put(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}).Produce(v1Type).Consume(v2Type)
			})
			It("should succeed get", func() {
				req, _ = http.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Accept", v1Type)
				gRouter.ServeHTTP(responseRecorder, req)
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})
			It("should select one accept header and set it", func() {
				req, _ = http.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Accept", fmt.Sprintf("%s; q=0.8, %s; q=1.0", v1Type, v2Type))
				gRouter.ServeHTTP(responseRecorder, req)
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				Expect(req.Header.Get("Accept")).To(Equal(v2Type))
			})
			It("should select the most precise media type", func() {
				req, _ = http.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Accept", fmt.Sprintf("%s, %s; q=1.0", v1Type, v2Type))
				gRouter.ServeHTTP(responseRecorder, req)
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				Expect(req.Header.Get("Accept")).To(Equal(v2Type))
			})
			It("should succeed post", func() {
				req, _ = http.NewRequest(http.MethodPost, "/", nil)
				req.Header.Set("Content-Type", v2Type)
				gRouter.ServeHTTP(responseRecorder, req)
				Expect(responseRecorder.Code).To(Equal(http.StatusCreated))
			})
			It("should succeed put", func() {
				req, _ = http.NewRequest(http.MethodPut, "/", nil)
				req.Header.Set("Accept", v1Type)
				req.Header.Set("Content-Type", v2Type)
				gRouter.ServeHTTP(responseRecorder, req)
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})
		})
		When("Defining a default route", func() {
			BeforeEach(func() {
				gRouter.Get(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}).Produce(v1Type, v2Type).SetDefault()
			})
			It("should succeed get", func() {
				req, _ = http.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Accept", "*/*")
				gRouter.ServeHTTP(responseRecorder, req)
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				Expect(req.Header.Get("Accept")).To(Equal(v1Type))

				req, _ = http.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Accept", "*")
				gRouter.ServeHTTP(responseRecorder, req)
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				Expect(req.Header.Get("Accept")).To(Equal(v1Type))
			})
		})
		When("sending wildcard content", func() {
			BeforeEach(func() {
				gRouter.Put(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusCreated)
				}).Consume(wildcardType, wildcardOnlyType, wildcardSlashType).Produce(v1Type).SetDefault()
			})
			It("should default the return content type", func() {
				req, _ = http.NewRequest(http.MethodPut, "/", nil)
				req.Header.Set("Content-Type", wildcardOnlyType)
				gRouter.ServeHTTP(responseRecorder, req)
				Expect(responseRecorder.Code).To(Equal(http.StatusCreated))
				Expect(req.Header.Get("Accept")).To(Equal(v1Type))

				req, _ = http.NewRequest(http.MethodPut, "/", nil)
				req.Header.Set("Content-Type", wildcardSlashType)
				gRouter.ServeHTTP(responseRecorder, req)
				Expect(responseRecorder.Code).To(Equal(http.StatusCreated))
				Expect(req.Header.Get("Accept")).To(Equal(v1Type))
			})
		})
		When("not providing accept or content-type headers", func() {
			BeforeEach(func() {
				gRouter.Get(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}).Produce(v1Type, v2Type)
				gRouter.Post(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusCreated)
				}).Consume(v1Type, v2Type)
				gRouter.Patch(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusCreated)
				}).Consume(v1Type, v2Type)
			})
			It("get should be unacceptable", func() {
				req, _ = http.NewRequest(http.MethodGet, "/", nil)
				gRouter.ServeHTTP(responseRecorder, req)
				Expect(responseRecorder.Code).To(Equal(http.StatusNotAcceptable))
			})
			It("post should negotiate", func() {
				req, _ = http.NewRequest(http.MethodPost, "/", nil)
				gRouter.ServeHTTP(responseRecorder, req)
				Expect(responseRecorder.Code).To(Equal(http.StatusUnsupportedMediaType))
				Expect(responseRecorder.Header().Get("Accept-Post")).To(Equal(fmt.Sprintf("%s, %s", v1Type, v2Type)))
			})
			It("patch should negotiate", func() {
				req, _ = http.NewRequest(http.MethodPatch, "/", nil)
				gRouter.ServeHTTP(responseRecorder, req)
				Expect(responseRecorder.Code).To(Equal(http.StatusUnsupportedMediaType))
				Expect(responseRecorder.Header().Get("Accept-Patch")).To(Equal(fmt.Sprintf("%s, %s", v1Type, v2Type)))
			})
		})
		When("providing unsupported headers", func() {
			BeforeEach(func() {
				gRouter.Get(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}).Produce(v1Type, v2Type)

				gRouter.Post(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusCreated)
				}).Consume(v1Type)

				gRouter.Post(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusCreated)
				}).Consume(v2Type)

				gRouter.Put(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}).Produce(v1Type).Consume(v1Type, v2Type)

				gRouter.Patch(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}).Produce(v1Type).Consume(v1Type, v2Type)
			})
			It("get should be not acceptable", func() {
				req, _ = http.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Accept", unsupportedType)
				gRouter.ServeHTTP(responseRecorder, req)
				Expect(responseRecorder.Code).To(Equal(http.StatusNotAcceptable))
			})
			It("put should be not acceptable", func() {
				req, _ = http.NewRequest(http.MethodPut, "/", nil)
				req.Header.Set("Accept", unsupportedType)
				gRouter.ServeHTTP(responseRecorder, req)
				Expect(responseRecorder.Code).To(Equal(http.StatusNotAcceptable))
			})
			It("post should negotiate", func() {
				req, _ = http.NewRequest(http.MethodPost, "/", nil)
				req.Header.Set("Content-Type", unsupportedType)
				gRouter.ServeHTTP(responseRecorder, req)
				Expect(responseRecorder.Code).To(Equal(http.StatusUnsupportedMediaType))
				Expect(responseRecorder.Header().Get("Accept-Post")).To(Equal(fmt.Sprintf("%s, %s", v1Type, v2Type)))
			})
			It("patch should negotiate", func() {
				req, _ = http.NewRequest(http.MethodPatch, "/", nil)
				req.Header.Set("Content-Type", unsupportedType)
				gRouter.ServeHTTP(responseRecorder, req)
				Expect(responseRecorder.Code).To(Equal(http.StatusUnsupportedMediaType))
				Expect(responseRecorder.Header().Get("Accept-Patch")).To(Equal(fmt.Sprintf("%s, %s", v1Type, v2Type)))
			})
		})
		When("providing an invalid headers", func() {
			BeforeEach(func() {
				gRouter.Get(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}).Produce(v1Type, v2Type)

				gRouter.Post(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusCreated)
				}).Consume(v1Type)

				gRouter.Post(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusCreated)
				}).Consume(v2Type)
			})
			It("should return 406", func() {
				req, _ = http.NewRequest(http.MethodPost, "/", nil)
				req.Header.Set("Content-Type", invalidType)
				gRouter.ServeHTTP(responseRecorder, req)
				Expect(responseRecorder.Code).To(Equal(http.StatusNotAcceptable))
			})
			It("should return 406", func() {
				req, _ = http.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Accept", invalidType)
				gRouter.ServeHTTP(responseRecorder, req)
				Expect(responseRecorder.Code).To(Equal(http.StatusNotAcceptable))
			})
		})
		When("requesting with bad http verb", func() {
			It("should return 405", func() {
				req, _ = http.NewRequest(http.MethodTrace, "/", nil)
				gRouter.ServeHTTP(responseRecorder, req)
				Expect(responseRecorder.Code).To(Equal(http.StatusMethodNotAllowed))
			})
		})
	})
})
