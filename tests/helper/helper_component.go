package helper

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

// WaitAppReadyInContainer probes the remote container using the specified command (cmd).
// It waits until the specified timeout is reached or until the provided matchers match the remote command output.
// At least one of the matchers must be provided.
func WaitAppReadyInContainer(
	cmp Component,
	container string,
	cmd []string,
	pollingInterval time.Duration,
	timeout time.Duration,
	stdoutMatcher types.GomegaMatcher,
	stderrMatcher types.GomegaMatcher,
) {
	if stdoutMatcher == nil && stderrMatcher == nil {
		Fail("Please specify either stdoutMatcher or stderrMatcher!")
	}
	Eventually(func(g Gomega) {
		stdout, stderr := cmp.Exec(container, cmd, nil)
		if stdoutMatcher != nil {
			g.Expect(stdout).Should(stdoutMatcher)
		}
		if stderrMatcher != nil {
			g.Expect(stderr).Should(stderrMatcher)
		}
	}).WithPolling(pollingInterval).WithTimeout(timeout).Should(Succeed())
}
