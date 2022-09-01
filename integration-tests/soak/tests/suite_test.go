<<<<<<< HEAD:integration-tests/soak/suite_test.go
package soak_test
=======
package tests_test
>>>>>>> 9147acd (Adding Soak support):integration-tests/soak/tests/suite_test.go

//revive:disable:dot-imports
import (
	"testing"

	"github.com/smartcontractkit/chainlink/integration-tests/actions"

	. "github.com/onsi/ginkgo/v2"
)

func Test_Suite(t *testing.T) {
	actions.GinkgoSuite()
	RunSpecs(t, "Integration")
}
