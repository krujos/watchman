package exceptionprocessor

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestExceptionprocessor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Exceptionprocessor Suite")
}
