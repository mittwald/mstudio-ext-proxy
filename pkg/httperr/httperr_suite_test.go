package httperr_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHttperr(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Httperr Suite")
}
