package middleware

import (
	"compress/flate"
	"compress/gzip"
	"container/heap"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Accept-Encoding: deflate, gzip;q=1.0, *;q=0.5

const defaultEncoding = "gzip"

var (
	supportedEncodings = map[string]compressFunc{
		"*":       compressWithGzip,
		"gzip":    compressWithGzip,
		"deflate": compressWithDeflate,
	}
)

func CompressResponse() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			supportedContentEncoding := getContentEncoding(r.Header.Get("Accept-Encoding"))
			cw := &compressWriter{w, supportedContentEncoding}
			cw.Header().Set("Content-Encoding", supportedContentEncoding)
			next.ServeHTTP(cw, r)
		})
	}
}

func getContentEncoding(acceptEncoding string) string {
	encodings := strings.Split(acceptEncoding, ",")
	qHeap := make(maxHeap, 0, len(encodings))
	for _, encoding := range encodings {
		compress, err := createQCompress(encoding)
		if err != nil {
			continue
		}
		heap.Push(&qHeap, compress)
	}
	for len(qHeap) > 0 {
		compress := heap.Pop(&qHeap).(qCompress)
		if supportedEncodings[compress.algo] != nil {
			if compress.algo == "*" {
				return defaultEncoding
			}
			return compress.algo
		}
	}
	return ""
}

func createQCompress(encoding string) (qCompress, error) {
	encoding = strings.TrimSpace(encoding)
	split := strings.Split(encoding, ";")
	compress := qCompress{algo: split[0], q: 1.0, qPresent: false}
	if len(split) > 1 {
		q := strings.TrimSpace(split[1])
		if strings.HasPrefix(q, "q=") {
			qFloat, err := strconv.ParseFloat(q[2:], 64)
			if err != nil {
				return compress, err
			}
			compress.q = qFloat
			compress.qPresent = true
		}
	}
	return compress, nil
}

// Wrapper to handle the compression of the response
type compressWriter struct {
	http.ResponseWriter
	acceptEncoding string
}

func (c *compressWriter) Write(b []byte) (int, error) {
	if supportedEncodings[c.acceptEncoding] != nil {
		return supportedEncodings[c.acceptEncoding](c.ResponseWriter, b)
	}
	return c.ResponseWriter.Write(b)
}

type compressFunc func(w http.ResponseWriter, b []byte) (int, error)

func compressWithGzip(w http.ResponseWriter, b []byte) (int, error) {
	zw := gzip.NewWriter(w)
	defer zw.Close()
	return zw.Write(b)
}

func compressWithDeflate(w http.ResponseWriter, b []byte) (int, error) {
	zw, err := flate.NewWriter(w, flate.DefaultCompression)
	if err != nil {
		return 0, errors.Wrap(err, "failed to create deflate writer")
	}
	defer zw.Close()
	return zw.Write(b)
}

type qCompress struct {
	q        float64
	algo     string
	qPresent bool
}

// Heap to sort the compressions by q value

type maxHeap []qCompress

func (h *maxHeap) Push(x any) {
	*h = append(*h, x.(qCompress))
}

func (h *maxHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h maxHeap) Len() int { return len(h) }
func (h maxHeap) Less(i, j int) bool {
	if h[i].q == h[j].q {
		return h[i].qPresent
	}
	return h[i].q > h[j].q
}
func (h maxHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
