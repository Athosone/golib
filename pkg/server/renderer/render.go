package renderer

import (
	"bytes"
	"container/heap"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"

	"github.com/athosone/golib/pkg/server"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var DefaultSerializer = "json"

var (
	supportedSerializer = map[string]serializer{
		"*":    renderJSON,
		"json": renderJSON,
		"yaml": renderYAML,
		"xml":  renderXML,
	}
)

type serializer func(http.ResponseWriter, any) (*bytes.Buffer, error)

// Helper function to render response

func OK(w http.ResponseWriter, r *http.Request, v any) error {
	return RenderResponse(w, r, http.StatusOK, v)
}

func Created(w http.ResponseWriter, r *http.Request, v any) error {
	return RenderResponse(w, r, http.StatusCreated, v)
}

func Accepted(w http.ResponseWriter, r *http.Request, v any) error {
	return RenderResponse(w, r, http.StatusAccepted, v)
}

func NotFound(w http.ResponseWriter, r *http.Request, v any) error {
	return RenderResponse(w, r, http.StatusNotFound, v)
}

func BadRequest(w http.ResponseWriter, r *http.Request, v any) error {
	return RenderResponse(w, r, http.StatusBadRequest, v)
}

// Encoding

// RenderResponse renders data passed to it encoded in the appropriate format.
// The format is determined by the Accept header.
// If the Accept header is not present, the default format is used.
// The DefaultContentType is used as default format, if you want to change it you can as its a var.
// If the Accept header is not supported, an error is returned.
func RenderResponse(w http.ResponseWriter, r *http.Request, status int, v any) error {
	buf, err := Encode(w, r, v)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	w.WriteHeader(status)
	_, err = w.Write(buf.Bytes())
	return err
}

func Encode(w http.ResponseWriter, r *http.Request, v any) (*bytes.Buffer, error) {
	accept := r.Header.Get("Accept")
	mediaType, err := searchContentType(accept)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to parse Accept header, invalid value: %s", accept))
	}
	if mediaType.IsAny || mediaType.Format == "*" {
		if DefaultSerializer == "" {
			return nil, errors.New("no Accept header and no DefaultSerializer defined")
		}
		mediaType.Format = DefaultSerializer
		if mediaType.IsAny {
			mediaType.FullyQualifiedType = "application/" + DefaultSerializer
		}
	}
	serializer, ok := supportedSerializer[mediaType.Format]
	if !ok {
		return nil, errors.Errorf("unsupported Accept header: %s", accept)
	}

	ft := strings.Replace(mediaType.FullyQualifiedType, "+*", "+"+mediaType.Format, 1)
	w.Header().Set("Content-Type", ft)
	return serializer(w, v)
}

func renderJSON(w http.ResponseWriter, data any) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(true)

	return &buf, encoder.Encode(data)
}

func renderYAML(w http.ResponseWriter, data any) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)

	return &buf, encoder.Encode(data)
}

func renderXML(w http.ResponseWriter, data any) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	encoder := xml.NewEncoder(&buf)

	return &buf, encoder.Encode(data)
}

// Media types

func searchContentType(accept string) (*server.ContentMediaType, error) {
	mh, err := server.ParseMediaType(accept)
	if err != nil {
		return nil, err
	}
	for len(mh) > 0 {
		mt := heap.Pop(&mh).(server.ContentMediaType)
		if supportedSerializer[mt.Format] != nil {
			return &mt, nil
		}
	}
	return nil, errors.New("no supported content type found")
}
