package pubsub

import (
	"context"
	"math/rand"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type testEvent struct{}

var _ = Describe("Queue events", func() {
	Context("Publishing an event to multiple receivers", func() {
		var rec1 <-chan testEvent
		var rec2 <-chan testEvent

		BeforeEach(func() {
			ctx := context.TODO()
			queue := NewTopic[testEvent](ctx)
			rec1 = queue.Subscribe()
			rec2 = queue.Subscribe()
			queue.Publish(testEvent{})
		})

		It("broadcasts the event", func() {
			Eventually(rec1).Should(Receive())
			Eventually(rec2).Should(Receive())
		})
	})
	Context("Cancelling the queue closes channels", func() {
		var receivers []<-chan testEvent
		ctx := context.TODO()
		var queue *Topic[testEvent]
		BeforeEach(func() {
			queue = NewTopic[testEvent](ctx)
			for i := 0; i < 50; i++ {
				rec := queue.Subscribe()
				go func(r <-chan testEvent) {
					for {
						select {
						case v := <-rec:
							Expect(v).ToNot(BeNil())
						case <-ctx.Done():
							return
						}
					}
				}(rec)
				receivers = append(receivers, queue.Subscribe())
			}
			rand.Seed(time.Now().UnixNano())
			for i := 0; i < 100; i++ {
				go func() {
					<-time.After(time.Duration(rand.Intn(50)) * time.Millisecond)
					queue.Publish(testEvent{})
				}()
			}
		})

		It("closes all of its receivers channel", func() {
			By("closing the queue")
			queue.Close()
			Expect(queue.IsClosed()).To(BeTrue())
			By("closing again does nothing")
			queue.Close()
			Expect(queue.IsClosed()).To(BeTrue())
		})
	})
})
