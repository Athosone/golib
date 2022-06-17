package server

import (
	"container/heap"
	"errors"
	"mime"
	"strconv"
	"strings"
)

// Based on accept header specified in: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept
func ParseMediaType(headerValue string) (MaxContentTypeHeap, error) {
	if headerValue == "" {
		return MaxContentTypeHeap{}, nil
	}
	var mh MaxContentTypeHeap
	for _, media := range strings.Split(headerValue, ",") {
		mediaType, params, err := mime.ParseMediaType(media)
		if err != nil {
			continue
		}

		_, isQualitySet := params["q"]
		qFloat := 1.0
		if isQualitySet {
			qFloat, err = strconv.ParseFloat(params["q"], 64)
			if err != nil {
				continue
			}
		}
		format := "*"
		formats := strings.Split(mediaType, "/")
		if len(formats) == 2 {
			format = formats[1]
		}
		if strings.Contains(format, "+") {
			splitStr := strings.Split(format, "+")
			if len(splitStr) > 2 {
				continue
			}
			format = splitStr[1]
		}
		isAny := mediaType == "*" || mediaType == "*/*"
		delete(params, "q")
		mt := ContentMediaType{mime.FormatMediaType(mediaType, params), mediaType, format, qFloat, isQualitySet, isAny}
		heap.Push(&mh, mt)
	}
	if len(mh) == 0 {
		return mh, errors.New("no supported content type found")
	}
	return mh, nil
}

type ContentMediaType struct {
	FullyQualifiedType, Type, Format string
	Quality                          float64
	IsQualitySet                     bool
	IsAny                            bool
}

// max heap on media type

type MaxContentTypeHeap []ContentMediaType

func (h MaxContentTypeHeap) Len() int { return len(h) }

func (h MaxContentTypeHeap) Less(i, j int) bool {
	if h[i].Quality == h[j].Quality {
		return h[i].IsQualitySet
	}
	return h[i].Quality > h[j].Quality
}

// Heap interface implementation
func (h MaxContentTypeHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *MaxContentTypeHeap) Push(x any) {
	*h = append(*h, x.(ContentMediaType))
}

func (h *MaxContentTypeHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
