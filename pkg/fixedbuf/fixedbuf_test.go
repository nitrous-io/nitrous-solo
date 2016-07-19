package fixedbuf_test

import (
	"testing"

	"github.com/nitrous-io/nitrous-solo/pkg/fixedbuf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "pkg/fixedbuf")
}

var _ = Describe("pkg/fixedbuf", func() {
	var b *fixedbuf.FixedBuf

	BeforeEach(func() {
		b = fixedbuf.New(10)
	})

	Context("writing to an empty buffer", func() {
		Context("when number of bytes written is less than the buffer size", func() {
			It("writes without discarding any data", func() {
				n, err := b.Write([]byte("hello"))
				Expect(err).To(BeNil())
				Expect(n).To(Equal(5))
				Expect(string(b.Bytes())).To(Equal("hello"))
			})
		})

		Context("when number of bytes written is equal to the buffer size", func() {
			It("writes without discarding any data", func() {
				n, err := b.Write([]byte("helloworld"))
				Expect(err).To(BeNil())
				Expect(n).To(Equal(10))
				Expect(string(b.Bytes())).To(Equal("helloworld"))
			})
		})

		Context("when number of bytes written is greater than the buffer size", func() {
			It("truncates and writes only the last n(=size) bytes", func() {
				n, err := b.Write([]byte("supermariobros!"))
				Expect(err).To(BeNil())
				Expect(n).To(Equal(15))
				Expect(string(b.Bytes())).To(Equal("mariobros!"))
			})
		})
	})

	Context("writing to a partially written buffer", func() {
		BeforeEach(func() {
			n, err := b.Write([]byte("hello"))
			Expect(err).To(BeNil())
			Expect(n).To(Equal(5))
		})

		Context("when number of bytes written is less than the free space available", func() {
			It("writes without discarding any data", func() {
				n, err := b.Write([]byte("boy"))
				Expect(err).To(BeNil())
				Expect(n).To(Equal(3))
				Expect(string(b.Bytes())).To(Equal("helloboy"))
			})
		})

		Context("when number of bytes written is equal to the free space available", func() {
			It("writes without discarding any data", func() {
				n, err := b.Write([]byte("world"))
				Expect(err).To(BeNil())
				Expect(n).To(Equal(5))
				Expect(string(b.Bytes())).To(Equal("helloworld"))
			})
		})

		Context("when number of bytes written is greater than the free space available", func() {
			It("discards some existing data and appends new data", func() {
				n, err := b.Write([]byte(" world"))
				Expect(err).To(BeNil())
				Expect(n).To(Equal(6))
				Expect(string(b.Bytes())).To(Equal("ello world"))
			})
		})

		Context("when number of bytes written is greater than the buffer size", func() {
			It("truncates and writes only the last n(=size) bytes", func() {
				n, err := b.Write([]byte("supermariobros!"))
				Expect(err).To(BeNil())
				Expect(n).To(Equal(15))
				Expect(string(b.Bytes())).To(Equal("mariobros!"))
			})
		})
	})

	Context("writing to a full buffer", func() {
		BeforeEach(func() {
			n, err := b.Write([]byte("helloworld"))
			Expect(err).To(BeNil())
			Expect(n).To(Equal(10))
		})

		Context("when number of bytes written is less than the buffer size", func() {
			It("discards some existing data and appends new data", func() {
				n, err := b.Write([]byte("hello"))
				Expect(err).To(BeNil())
				Expect(n).To(Equal(5))
				Expect(string(b.Bytes())).To(Equal("worldhello"))
			})
		})

		Context("when number of bytes written is equal to the buffer size", func() {
			It("discards all existing data and writes new data", func() {
				n, err := b.Write([]byte("good bye!!"))
				Expect(err).To(BeNil())
				Expect(n).To(Equal(10))
				Expect(string(b.Bytes())).To(Equal("good bye!!"))
			})
		})

		Context("when number of bytes written is greater than the buffer size", func() {
			It("truncates and writes only the last n(=size) bytes of the new data", func() {
				n, err := b.Write([]byte("supermariobros!"))
				Expect(err).To(BeNil())
				Expect(n).To(Equal(15))
				Expect(string(b.Bytes())).To(Equal("mariobros!"))
			})
		})
	})
})
