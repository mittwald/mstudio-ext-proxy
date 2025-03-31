package httperr_test

import (
	"fmt"
	httperr2 "github.com/mittwald/api-client-go/pkg/httperr"
	"github.com/mittwald/mstudio-ext-proxy/pkg/httperr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"net/http"
)

var _ = Describe("StatusForError", func() {
	It("should correctly extract status from wrapped error", func() {
		inner := fmt.Errorf("foo")
		err := httperr.ErrWithStatus(http.StatusPaymentRequired, "payment required", inner)

		Expect(httperr.StatusForError(err)).To(Equal(http.StatusPaymentRequired))
	})

	It("should correctly infer status from mittwald client errors", func() {
		err := &httperr2.ErrNotFound{}
		Expect(httperr.StatusForError(err)).To(Equal(http.StatusNotFound))
	})

	It("should fall back to 500", func() {
		err := fmt.Errorf("foo")

		Expect(httperr.StatusForError(err)).To(Equal(http.StatusInternalServerError))
	})
})
