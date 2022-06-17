package middleware_test

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/athosone/golib/pkg/server/middleware"
)

type testHandler func(http.ResponseWriter, *http.Request)

var _ = Describe("Compress", func() {
	var (
		compressMiddleware func(http.Handler) http.Handler
		handlerFunc        http.Handler
		request            *http.Request
		responseRecorder   *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		compressMiddleware = middleware.CompressResponse()
		handlerFunc = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("test"))
		})
		request, _ = http.NewRequest("GET", "/", nil)
		responseRecorder = httptest.NewRecorder()
	})

	When("Request contains a single choice of encoding", func() {
		BeforeEach(func() {
			request.Header.Set("Accept-Encoding", "gzip")
		})
		It("should compress with requested encoding", func() {
			compressMiddleware(handlerFunc).ServeHTTP(responseRecorder, request)
			Expect(responseRecorder.Header().Get("Content-Encoding")).To(Equal("gzip"))
		})
	})
	When("Request specify any encoding *", func() {
		BeforeEach(func() {
			request.Header.Set("Accept-Encoding", "*")
		})
		It("should compress with gzip encoding", func() {
			compressMiddleware(handlerFunc).ServeHTTP(responseRecorder, request)
			Expect(responseRecorder.Header().Get("Content-Encoding")).To(Equal("gzip"))
		})
	})

	When("Encoding is not supported", func() {
		BeforeEach(func() {
			request.Header.Set("Accept-Encoding", "br")
		})
		It("should not compress", func() {
			compressMiddleware(handlerFunc).ServeHTTP(responseRecorder, request)
			Expect(responseRecorder.Header().Get("Content-Encoding")).To(Equal(""))
		})
	})
	When("Request contains multiple choice of encoding with different quality", func() {
		BeforeEach(func() {
			request.Header.Set("Accept-Encoding", "deflate;q=1.0, gzip;q=0.5")
		})
		It("should encode with the highest quality", func() {
			compressMiddleware(handlerFunc).ServeHTTP(responseRecorder, request)
			Expect(responseRecorder.Header().Get("Content-Encoding")).To(Equal("deflate"))
		})
	})
	When("Request contains multiple choice of encoding with the same quality", func() {
		BeforeEach(func() {
			request.Header.Set("Accept-Encoding", "gzip;q=1.0, deflate")
		})
		It("should encode with the highest and most precise quality", func() {
			compressMiddleware(handlerFunc).ServeHTTP(responseRecorder, request)
			Expect(responseRecorder.Header().Get("Content-Encoding")).To(Equal("gzip"))
		})
	})

	Context("Testing that content body is compressed", func() {
		When("Request asks for gzip compression", func() {
			BeforeEach(func() {
				request.Header.Set("Accept-Encoding", "gzip")
			})
			It("should compress the body with gzip algorithm", func() {
				compressMiddleware(handlerFunc).ServeHTTP(responseRecorder, request)
				reader, err := gzip.NewReader(responseRecorder.Body)
				Expect(err).To(BeNil())
				defer reader.Close()
				// unzip the body with gzip reader
				got := make([]byte, len(responseRecorder.Body.Bytes()))
				n, err := reader.Read(got)
				if err != nil && err != io.EOF {
					Expect(err).To(BeNil())
				}
				// check if the body is the same as the original
				Expect(string(got[:n])).To(Equal("test"))
			})
		})
		When("Request asks for deflate compression", func() {
			BeforeEach(func() {
				request.Header.Set("Accept-Encoding", "deflate")
			})
			It("should compress the body with deflate algorithm", func() {
				compressMiddleware(handlerFunc).ServeHTTP(responseRecorder, request)
				reader := flate.NewReader(responseRecorder.Body)
				defer reader.Close()
				// unzip the body with gzip reader
				got := make([]byte, len(responseRecorder.Body.Bytes()))
				n, err := reader.Read(got)
				if err != nil && err != io.EOF {
					Expect(err).To(BeNil())
				}
				// check if the body is the same as the original
				Expect(string(got[:n])).To(Equal("test"))
			})
		})
	})
})
