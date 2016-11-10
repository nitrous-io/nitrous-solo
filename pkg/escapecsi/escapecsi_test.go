package escapecsi_test

import (
	"testing"

	"github.com/nitrous-io/nitrous-solo/pkg/escapecsi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "escapecsi")
}

var _ = Describe("EscapeCSI", func() {
	It("removes ansi escape code", func() {
		Expect(escapecsi.EscapeCSI([]byte{27, '[', '4', ';', '5', 'H'})).To(Equal([]byte{}))
		Expect(escapecsi.EscapeCSI([]byte{'B', 27, '[', '4', ';', '5', 'H', 'A'})).To(Equal([]byte{'B', 'A'}))
		Expect(escapecsi.EscapeCSI([]byte{'B', 27, '[', 'A', ';', '5', 'H', 'A'})).To(Equal([]byte{'B', 27, '[', 'A', ';', '5', 'H', 'A'}))
		Expect(escapecsi.EscapeCSI([]byte{'B', 27, '[', '4', '5', ';', '5', '7', 'H', 'A'})).To(Equal([]byte{'B', 'A'}))
		Expect(escapecsi.EscapeCSI([]byte{'B', 27, '[', '4', '5', 'A', 'A'})).To(Equal([]byte{'B', 'A'}))
	})

	It("does not remove color code", func() {
		Expect(escapecsi.EscapeCSI([]byte{27, '[', '4', ';', '5', 'm'})).To(Equal([]byte{27, '[', '4', ';', '5', 'm'}))
	})

	It("does not remove non ansi escape code", func() {
		Expect(escapecsi.EscapeCSI([]byte{'A', 'B', 'C', 'D'})).To(Equal([]byte{'A', 'B', 'C', 'D'}))
	})
})
