package gstorage_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGstorage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gstorage Suite")
}
