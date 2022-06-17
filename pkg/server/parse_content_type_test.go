package server_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/athosone/golib/pkg/server"
)

var _ = Describe("ParseContentType", func() {

	It("Simple content type", func() {
		contentType := "application/json"
		expected := server.ContentMediaType{
			FullyQualifiedType: "application/json",
			Type:               "application/json",
			Format:             "json",
			Quality:            1.0,
			IsQualitySet:       false,
		}
		actual, err := server.ParseMediaType(contentType)
		Expect(err).To(BeNil())
		Expect(actual[0]).To(Equal(expected))
	})

	It("Simple content type with quality", func() {
		contentType := "application/json;q=0.5"
		expected := server.ContentMediaType{
			FullyQualifiedType: "application/json",
			Type:               "application/json",
			Format:             "json",
			Quality:            0.5,
			IsQualitySet:       true,
		}
		actual, err := server.ParseMediaType(contentType)
		Expect(err).To(BeNil())
		Expect(actual[0]).To(Equal(expected))
	})

	It("Orders content type by quality", func() {
		contentType := "application/json;q=0.5, application/xml;q=0.8"
		expected := server.ContentMediaType{
			FullyQualifiedType: "application/xml",
			Type:               "application/xml",
			Format:             "xml",
			Quality:            0.8,
			IsQualitySet:       true,
		}
		actual, err := server.ParseMediaType(contentType)
		Expect(err).To(BeNil())
		Expect(actual[0]).To(Equal(expected))
	})

	It("Orders content type by quality and type", func() {
		contentType := "application/yaml, application/json;q=1"
		expected := server.ContentMediaType{
			FullyQualifiedType: "application/json",
			Type:               "application/json",
			Format:             "json",
			Quality:            1.0,
			IsQualitySet:       true,
		}
		actual, err := server.ParseMediaType(contentType)
		Expect(err).To(BeNil())
		Expect(actual[0]).To(Equal(expected))
	})

	It("Parse custom content type", func() {
		contentType := "application/vnd.athosonecentral.innersource+json;     v=1"
		expected := server.ContentMediaType{
			FullyQualifiedType: "application/vnd.athosonecentral.innersource+json; v=1",
			Type:               "application/vnd.athosonecentral.innersource+json",
			Format:             "json",
			Quality:            1.0,
			IsQualitySet:       false,
		}
		actual, err := server.ParseMediaType(contentType)
		Expect(err).To(BeNil())
		Expect(actual[0]).To(Equal(expected))
	})
})
