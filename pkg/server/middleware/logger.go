package middleware

import (
	"fmt"
	"net/http"
	"time"

	glogger "github.com/athosone/golib/pkg/logger"
	"go.uber.org/zap"
)

type responseWrapper struct {
	http.ResponseWriter
	status int
}

func (resp *responseWrapper) WriteHeader(status int) {
	resp.status = status
	resp.ResponseWriter.WriteHeader(status)
}

func RequestLogger(excludedPath []string) func(next http.Handler) http.Handler {
	excludedPathMap := make(map[string]struct{}, len(excludedPath))
	for _, p := range excludedPath {
		excludedPathMap[p] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := glogger.LoggerFromContextOrDefault(r.Context())
			requestTime := time.Now()
			rw := &responseWrapper{w, http.StatusOK}
			next.ServeHTTP(rw, r)

			if logger == nil {
				return
			}
			if _, ok := excludedPathMap[r.URL.Path]; ok {
				return
			}
			logger.Infow("",
				"method", r.Method,
				"url", fmt.Sprintf("%s%s", r.Host, r.URL.String()),
				"protocol", r.Proto,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
				"elapsed_time", time.Since(requestTime).String(),
				"request_headers", r.Header,
				"response_headers", rw.Header(),
				"response_status", rw.status,
			)
		})
	}
}

type LoggerFactory func(*http.Request) *zap.SugaredLogger

// InjectLoggerInRequest injects a logger into the request context
// The logger can then be retrieved from the request context using the logger.LoggerFromContextOrDefault(context.Context) function in the logger package.
// The logger factory param is used to create a logger for each request. In the factory you could specify which values you want to pass to subsequent middleware/controllers.
func InjectLoggerInRequest(logFactory LoggerFactory) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := logFactory(r)
			ctx := glogger.NewContextWithLogger(r.Context(), l)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
