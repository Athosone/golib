# Introduction

This repo is a set of *independent* packages that are used to build go applications.

It provides a set of tools that are used to build RESTful APIs, setup configuration, logger, provides pubsub mechanism.

They includes:

- [auth](pkg/auth/)

- [config](pkg/config/)

- [logger](pkg/logger/)

- [pubsub](pkg/pubsub/)

- [utils](pkg/utils/)

- [server](pkg/server/)
  - [middleware](pkg/server/middleware/)
  - [renderer](pkg/server/renderer/)
  - [routing](pkg/server/routing/)


## Prerequisites

- [golang v1.18+](https://golang.org/doc/install)
- To run unit tests: [ginkgo](https://onsi.github.io/ginkgo/)
  - `go install -mod=mod github.com/onsi/ginkgo/v2/ginkgo@latest`

# Package description

## config

The config package is useful to setup configuration for your application.

It uses the [viper library](https://github.com/spf13/viper).

Using this package you can load a strongly typed configuration from a file, environment variables, command line arguments, etc.

You can check the unit tests for more examples.

## logger

The logger package is used to wrap the `zap` library.

It includes a few helper methods to inject and retrieve logger from the `context.Context`.

The [zap logger](https://github.com/uber-go/zap) is a highly configurable and performant logging library for Go.

It is also very popular in the community.

Notably it allows for structured logging and the format can easily be customized.

## pubsub

The pubsub package is used to publish and subscribe messages in memory.

It leverage go channels to enable a pub sub mechanism which can be useful when you want to `fan-out` an event to multiple receivers.

The package is thread safe.

## utils

Every project has a trash folder and here it's `utils`.

`utils` package is used to put methods that are used in multiple packages but don't have a clear boundary like the method `GenerateRandomNameWithPrefix`.

## server

This package contains many components that can be used to build a web server.

### [serve](pkg/server/serve.go) is the main entry point to start a server

The server function starts a server and waits for the following signals:

- SIGINT
- SIGTERM
- SIGHUP
- SIGQUIT

Upon receiving any of these signals the server will gracefully shutdown.

It will call the `Shutdown` method of the `Server` interface passing a deadline of `5s` before definitively exiting the program.

### [server/middleware](pkg/server/middleware)

This is a collection of middlewares that can be used with net/http compliant servers.

#### [compress](pkg/server/middleware/compress.go)

The compress middleware is used to compress the response based on the `Accept-Encoding` header.

The middleware supports multiple encodings and will compress the response based on the quality parameter following the spec: [Accept-Encoding](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept-Encoding)

Sets the response `Content-Encoding` header.

At the moment only `gzip` and `deflate` encodings are supported.

#### [logger](pkg/server/middleware/logger.go)

There are two middleware in this package.

The first one is the [InjectLoggerInRequest](pkg/server/middleware/logger.go#InjectLoggerInRequest).

This function will inject a logger in the `context.Context` of the request using [NewContextWithLogger](pkg/logger/logger.go#NewContextWithLogger) from the logger package.

Subsequent middleware/handlers will be able to retrieve the logger using [LoggerFromContextOrDefault](pkg/logger/logger.go#LoggerFromContextOrDefault) from the logger package.

The other middleware is the [RequestLogger](pkg/server/middleware/logger.go#RequestLogger).

It is used to log incomings requests and the responses.

In order for RequestLogger to work you have to use the `InjectLoggerInRequest` first.

### [renderer](pkg/server/renderer/render.go)

The renderer package is used to render the response based on the `Accept` header.

It follows the spec: [Accept](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept)

As such it is able to sort the supported media types by quality and render the response based on the most precise match.

Then the format is extracted from the media type and the data is serialized.

It will set the `Content-Type` header to the selected media type from the `Accept` header.

At the moment it only supports `yaml` `json` and `xml` serializer.

If `Accept` header is set to `*/*` or `*`, it will render the response as defined by the [DefaultSerializer](pkg/server/renderer/render.go#DefaultSerializer).

### [routing](pkg/server/routing/routing.go)

The routing package is used to build a router based on `MediaType` versioning.

This package allows you to define routes based on methods and media types. You can set a default route per method (i.e.: if you have multiple handlers for GET).

If your route accepts wildcard media types, the router will choose the first entry defined in the route.

The router has no dependency on any external package and can be plugged-in easily in any famous framework ([go-chi](https://go-chi.io/#/), [gorilla-mux](https://github.com/gorilla/mux)).

Check the [examples](examples/media-type-versioning/books/controller.go#BookingRouter) to see how to use it.

The router respects the spec: [Content-negotiation](https://developer.mozilla.org/en-US/docs/Web/HTTP/Content_negotiation)

## Misc

https://developer.mozilla.org/en-US/docs/Glossary/Quality_values

https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types