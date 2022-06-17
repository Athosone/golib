package routing

import (
	"container/heap"
	"fmt"
	"mime"
	"net/http"
	"strings"

	"github.com/athosone/golib/pkg/server"
	"github.com/athosone/golib/pkg/utils"
	"go.uber.org/zap"
)

const (
	// HeaderAccept is the header key for the Accept header
	HeaderAccept      = "Accept"
	HeaderContentType = "Content-Type"
)

func (gr *GRouter) Post(dest http.HandlerFunc) *Route {
	route := &Route{
		id:     utils.RandomString(30),
		dest:   dest,
		method: http.MethodPost,
	}
	gr.routes = append(gr.routes, route)
	return route
}

func (gr *GRouter) Get(dest http.HandlerFunc) *Route {
	route := &Route{
		id:     utils.RandomString(30),
		dest:   dest,
		method: http.MethodGet,
	}
	gr.routes = append(gr.routes, route)
	return route
}

func (gr *GRouter) Put(dest http.HandlerFunc) *Route {
	route := &Route{
		id:     utils.RandomString(30),
		dest:   dest,
		method: http.MethodPut,
	}
	gr.routes = append(gr.routes, route)
	return route
}

func (gr *GRouter) Delete(dest http.HandlerFunc) *Route {
	route := &Route{
		id:     utils.RandomString(30),
		dest:   dest,
		method: http.MethodDelete,
	}
	gr.routes = append(gr.routes, route)
	return route
}

func (gr *GRouter) Patch(dest http.HandlerFunc) *Route {
	route := &Route{
		id:     utils.RandomString(30),
		dest:   dest,
		method: http.MethodPatch,
	}
	gr.routes = append(gr.routes, route)
	return route
}

func (gr *GRouter) Head(dest http.HandlerFunc) *Route {
	route := &Route{
		id:     utils.RandomString(30),
		dest:   dest,
		method: http.MethodHead,
	}
	gr.routes = append(gr.routes, route)
	return route
}

func (gr *GRouter) Options(dest http.HandlerFunc) *Route {
	route := &Route{
		id:     utils.RandomString(30),
		dest:   dest,
		method: http.MethodOptions,
	}
	gr.routes = append(gr.routes, route)
	return route
}

func (r *Route) Consume(mediaTypes ...string) *Route {
	for _, mediaType := range mediaTypes {
		formattedMediaType := formatMediaType(mediaType)
		r.consume = append(r.consume, NewPattern(formattedMediaType))
	}
	return r
}

// Define media types produced by the route
// The first media type is the default one in case of no Accept header / wildcard
func (r *Route) Produce(mediaTypes ...string) *Route {
	for _, mediaType := range mediaTypes {
		formattedMediaType := formatMediaType(mediaType)
		r.produce = append(r.produce, NewPattern(formattedMediaType))
	}
	return r
}

func formatMediaType(mediaType string) string {
	if mediaType == "" {
		return ""
	}
	mt, params, err := mime.ParseMediaType(mediaType)
	if err != nil {
		zap.S().Fatalw("Invalid route configuration, unable to parse media type", "mediaType", mediaType)
	}
	delete(params, "charset")
	delete(params, "q")
	mt = mime.FormatMediaType(mt, params)
	return mt
}

type Patterns []Pattern

type Route struct {
	id        string
	method    string
	isDefault bool
	dest      http.HandlerFunc
	consume   Patterns
	produce   Patterns
}

type GRouter struct {
	routes []*Route
}

func NewRouter() *GRouter {
	return &GRouter{
		routes: []*Route{},
	}
}

func (gr *GRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	accept := r.Header.Get(HeaderAccept)
	contentType := r.Header.Get(HeaderContentType)

	acceptHeap, err := server.ParseMediaType(accept)
	if err != nil {
		// if invalid accept header, return 406
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	consumeHeap, err := server.ParseMediaType(contentType)
	if err != nil {
		// if invalid content-type header, return 406
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}
	routes := []*Route{}
	for _, route := range gr.routes {
		if route.isMethodMatch(r.Method) {
			routes = append(routes, route)
		}
	}

	if len(routes) == 0 {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	mRouter := map[string]*Route{}
	for len(consumeHeap) > 0 {
		consumeContentType := heap.Pop(&consumeHeap).(server.ContentMediaType)
		for _, route := range routes {
			if route.IsConsumeMatch(r, consumeContentType) {
				if len(acceptHeap) == 0 || acceptHeap[0].IsAny {
					if len(route.produce) > 0 {
						r.Header.Set(HeaderAccept, route.produce[0].getFullyQualifiedType())
					}
					route.ServeHTTP(w, r)
					return
				}
				mRouter[route.id] = route
			}
		}
	}
	if contentType != "" && len(mRouter) == 0 {
		gr.negotiate(w, r)
		return
	}

	for len(acceptHeap) > 0 {
		acceptContentType := heap.Pop(&acceptHeap).(server.ContentMediaType)
		for _, route := range routes {
			if route.IsProduceMatch(r, acceptContentType) {
				_, hasRoutePair := mRouter[route.id]
				if contentType == "" || hasRoutePair {
					r.Header.Set(HeaderAccept, acceptContentType.FullyQualifiedType)
					if acceptContentType.IsAny && len(route.produce) > 0 {
						r.Header.Set(HeaderAccept, route.produce[0].getFullyQualifiedType())
					}
					route.ServeHTTP(w, r)
					return
				}
			}
		}
	}
	gr.negotiate(w, r)
}

func (r *Route) IsProduceMatch(req *http.Request, value server.ContentMediaType) bool {
	return r.isMatch(req, value, r.produce)
}

func (r *Route) IsConsumeMatch(req *http.Request, value server.ContentMediaType) bool {
	return r.isMatch(req, value, r.consume)
}

func (r *Route) SetDefault() {
	r.isDefault = true
}

func (ro *Route) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ro.dest.ServeHTTP(w, r)
}

func (r *Route) isMatch(req *http.Request, value server.ContentMediaType, patterns Patterns) bool {
	if r.isDefault && value.IsAny {
		return r.isMethodMatch(req.Method)
	}
	return !value.IsAny && match(value.FullyQualifiedType, patterns)
}

func match(value string, patterns []Pattern) bool {
	if len(patterns) == 0 {
		return false
	}
	for _, pattern := range patterns {
		if pattern.Match(value) {
			return true
		}
	}
	return false
}

func (r *Route) isMethodMatch(method string) bool {
	mr := strings.ToUpper(method)
	mm := strings.ToUpper(r.method)
	return mm == mr
}

func (gr *GRouter) negotiate(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost && req.Method != http.MethodPatch {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}
	var b strings.Builder
	for _, route := range gr.routes {
		if route.isMethodMatch(req.Method) {
			var supportedMediaTypes string
			if len(route.consume) > 0 {
				supportedMediaTypes = route.getSupportedConsumeTypes()
			} else {
				supportedMediaTypes = route.getSupportedProduceTypes()
			}
			if b.Len() > 0 {
				b.WriteString(", ")
			}
			b.WriteString(supportedMediaTypes)
		}
	}
	w.Header().Set(fmt.Sprintf("Accept-%s", req.Method), b.String())
	w.WriteHeader(http.StatusUnsupportedMediaType)
}

func (r *Route) getSupportedConsumeTypes() string {
	return r.consume.concatenate()
}
func (r *Route) getSupportedProduceTypes() string {
	return r.produce.concatenate()
}

func (p Patterns) concatenate() string {
	var types []string
	for _, pattern := range p {
		types = append(types, pattern.getFullyQualifiedType())
	}
	return strings.Join(types, ", ")
}

func (p Pattern) getFullyQualifiedType() string {
	var b strings.Builder
	b.WriteString(p.prefix)
	if p.wildcard {
		b.WriteString("*")
	}
	b.WriteString(p.suffix)
	return b.String()
}
