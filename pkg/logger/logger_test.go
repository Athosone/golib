package logger_test

import (
	"context"

	glogger "github.com/athosone/golib/pkg/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

var _ = Describe("Logger", func() {
	var (
		logger *zap.SugaredLogger = glogger.NewLogger(false)
		ctx    context.Context
	)

	When("Adding logger to context", func() {
		BeforeEach(func() {
			ctx = glogger.NewContextWithLogger(context.Background(), logger)
		})
		It("should add logger to context", func() {
			Expect(ctx.Value(glogger.LoggerContextKey)).To(Equal(logger))
		})
		It("should fetch logger from context", func() {
			fetched := glogger.LoggerFromContextOrDefault(ctx)
			Expect(fetched).To(Equal(logger))
		})
	})
})
