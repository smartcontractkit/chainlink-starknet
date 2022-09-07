//revive:disable:dot-imports
package soak_test

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/smartcontractkit/chainlink/integration-tests/actions"
	"testing"
)

func Test_Suite(t *testing.T) {
	actions.GinkgoSuite()
	RunSpecs(t, "Integration")
}
